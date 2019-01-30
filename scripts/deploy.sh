#!/usr/bin/env bash
set -ueo pipefail
# set -x

ME=$(basename "$0")
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
GPG_PRIVATE_KEY="$PROJECT_ROOT/sonar-agent.key"
DOCKER_IMAGE="docker.io/digitalocean/do-agent"
VERSION=${VERSION:-$(cat target/VERSION || true)}
VERSION_REGEX="[^\\d]${VERSION}[^\\d]"

FORCE_RELEASE=${FORCE_RELEASE:-0}
SKIP_BACKUP=${SKIP_BACKUP:-0}
SKIP_CLEANUP=${SKIP_CLEANUP:-0}

CI_LOG_URL=""
[ -n "${CI_BASE_URL:-}" ] && CI_LOG_URL="${CI_BASE_URL}/tab/build/detail/${GO_PIPELINE_NAME}/${GO_PIPELINE_COUNTER}/${GO_STAGE_NAME}/${GO_STAGE_COUNTER}/${GO_JOB_NAME}"

# display usage for this script
function usage() {
	cat <<-EOF
	NAME:
	
	  $ME
	
	SYNOPSIS:

	  $ME <cmd>

	DESCRIPTION:

	    The purpose of this script is to publish build artifacts
	    to Github, docker.io, and apt/yum repositories.

	ENVIRONMENT:
		
	    VERSION (required)
	        The version to publish

	    GITHUB_AUTH_USER, GITHUB_AUTH_TOKEN (required)
	        Github access credentials

	    DOCKER_USER, DOCKER_PASSWORD  (required)
	        Docker hub access credentials

	    SPACES_ACCESS_KEY_ID, SPACES_SECRET_ACCESS_KEY (required)
	        DigitalOcean Spaces access credentials

	    SLACK_WEBHOOK_URL (optional)
	        Webhook URL to send notifications. Enables Slack
	        notifications

	COMMANDS:
	
	    github
	        push target/ assets to github

	    docker
	        push docker builds to hub.docker.io

	    spaces
	        publish artifacts to DigitalOcean Spaces

	    all
	        deploy all builds

	EOF
}

function main() {
	cmd=${1:-}

	case "$cmd" in
		github)
			check_version
			check_target_files
			cleanup
			deploy_github
			;;
		spaces)
			check_version
			check_target_files
			cleanup
			deploy_spaces
			;;
		docker)
			check_version
			check_target_files
			cleanup
			deploy_docker
			;;
		promote)
			check_version
			cleanup
			promote_spaces
			promote_github
			promote_docker
			;;
		all)
			check_version
			check_target_files
			cleanup
			deploy_spaces
			deploy_github
			deploy_docker
			;;
		help|--help|-h)
			usage
			exit 0
			;;
		*)
			abort "Unknown command '$cmd'. See $ME --help for help"
			;;
	esac
}

# verify the VERSION env var
function check_version() {
	[[ "${VERSION:-}" =~ [0-9]+\.[0-9]+\.[0-9]+ ]] || \
		abort "VERSION is required and should be semver format (e.g. 1.2.34)"
}

function verify_gpg_key() {
	stat "$GPG_PRIVATE_KEY" > /dev/null || abort "$GPG_PRIVATE_KEY is required"
}

function force_release_enabled() {
	if is_enabled "${FORCE_RELEASE}" ; then
		echo
		echo "WARNING! forcing a release of $VERSION"
		echo
		return 0
	fi
	return 1
}

function cleanup() {
	if is_enabled "${SKIP_CLEANUP}" ; then
		anounce "SKIP_CLEANUP is set to ${SKIP_CLEANUP}, skipping this step"
		return
	fi
	rm -rf "${PROJECT_ROOT}/repos"
}

# if a release with the VERSION tag is already published then we cannot deploy
# this version over the previous release.
function check_can_deploy_spaces() {
	force_release_enabled && return 0
	anounce "Checking if we can deploy spaces"

	status_code=$(http_status_for "https://insights.nyc3.digitaloceanspaces.com/apt-beta/pool/main/main/d/do-agent/do-agent_${VERSION}_amd64.deb")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been deployed. Add a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

function deploy_spaces() {
	check_can_deploy_spaces
	pull_spaces /
	backup_spaces

	anounce "Deploying packages to spaces"
	target_files | grep -P '\.deb$' | while IFS= read -r file; do
		cp -Luv "$file" repos/apt-beta/pool/main/main/d/do-agent/
	done

	target_files | grep -P '\.rpm$' | while IFS= read -r file; do
		dest=repos/yum-beta/x86_64/
		[[ "$file" =~ "i386" ]] && \
			dest=repos/yum-beta/i386/
		cp -Luv "$file" "$dest"
	done

	rebuild_apt_beta_packages
	rebuild_yum_beta_packages

	# sync the packages first to prevent race conditions
	push_spaces "/apt-beta/pool/main/" --exclude "*" --include "**/*.deb"
	push_spaces "/yum-beta/" --exclude "*" --include "*/*.rpm"

	# then sync the metadata and everything else
	push_spaces "/apt-beta/dists/main/"
	push_spaces "/yum-beta/"

	anounce "Deploy spaces completed"
}

function rebuild_apt_main_packages() {
	verify_gpg_key
	anounce "Rebuilding apt package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/apt:/work/apt" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-apt:5b0c797" \
		|| abort "Failed to rebuild apt package indexes"
}

function rebuild_apt_beta_packages() {
	verify_gpg_key
	anounce "Rebuilding apt package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/apt-beta:/work/apt" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-apt:5b0c797" \
		|| abort "Failed to rebuild apt package indexes"
}

function rebuild_yum_main_packages() {
	verify_gpg_key
	anounce "Rebuilding yum package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/yum:/work/yum" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-yum:5b0c797" \
		|| abort "Failed to rebuild yum package indexes"
}

function rebuild_yum_beta_packages() {
	verify_gpg_key
	anounce "Rebuilding yum package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/yum-beta:/work/yum" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-yum:5b0c797" \
		|| abort "Failed to rebuild yum package indexes"
}

# sync local cache to Spaces
#
# Usage: push_spaces <path> [optional aws cli args]
#
# Examples:
#    # push everything
#    push_spaces /
#    push_spaces /apt-beta/pool/main --include "*" --exclude "*.txt"
#    push_spaces / --include "*" --exclude "*.txt"
function push_spaces() {
	path=${1:-}
	[ -z "$path" ] && abort "Usage: ${FUNCNAME[0]} <path> [optional aws cli args]"
	[[ ! "$path" =~ ^/ ]] && abort "<path> must begin with a slash"

	anounce "Syncing Spaces changes to to ${path}"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		"./repos${path}" \
		"s3://insights${path}" \
		--delete \
		--acl public-read \
		"${@:2}"
}

# sync Spaces directory to local cache
#
# Usage: pull_spaces <path> [optional aws cli args]
#
# Examples:
#    # pull everything
#    pull_spaces /
#    pull_spaces /apt-beta/pool/main --include "*" --exclude "*.txt"
#    pull_spaces / --include "*" --exclude "*.txt"
function pull_spaces() {
	path=${1:-}
	[ -z "$path" ] && abort "Usage: ${FUNCNAME[0]} <path> [optional aws cli args]"
	[[ ! "$path" =~ ^/ ]] && abort "<path> must begin with a slash"

	anounce "Syncing Spaces to local cache"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		"s3://insights${path}" \
		"./repos${path}" \
		--delete \
		--acl public-read \
		"${@:2}"
}

# back up the local ./repos/ directory to ams3 Space "insights2"
# used to keep backups before deploying
function backup_spaces() {
	anounce "Backing up Spaces"
	if is_enabled "${SKIP_BACKUP}" ; then
		anounce "SKIP_BACKUP is set to ${SKIP_BACKUP}, skipping..."
		return
	fi
	pull_spaces /

	aws s3 \
		--endpoint-url https://ams3.digitaloceanspaces.com \
		sync \
		./repos/ \
		s3://insights2/ \
		--delete \
		--acl public-read
}

# interact with the awscli via docker
function aws() {
	docker run \
		--rm -i \
		-e "AWS_ACCESS_KEY_ID=${SPACES_ACCESS_KEY_ID}" \
		-e "AWS_SECRET_ACCESS_KEY=${SPACES_SECRET_ACCESS_KEY}" \
		-e "AWS_DEFAULT_REGION=nyc3" \
		-v "$PROJECT_ROOT:/project" \
		-w /project \
		-u "$(id -u)" \
		mesosphere/aws-cli \
		"$@"
}

function check_can_deploy_github() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://api.github.com/repos/digitalocean/do-agent/releases/tags/${VERSION}")

	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been released. Add a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}


# deploy the compiled binaries and packages to github releases
function deploy_github() {
	check_can_deploy_github
	anounce "Deploying to Github"

	create_github_release || abort "Github deploy failed"

	upload_url=$(github_asset_upload_url)

	for file in $(target_files); do
		name=$(basename "$file")

		echo "Uploading $name to github"
		github_curl \
			-X "POST" \
			-H "Content-Type: $(content_type_for "$file")" \
			--data-binary "@${file}" \
			"$upload_url?name=$name" \
			| jq -r '. | "Success: \(.name)"' &
	done
	wait
}

function check_can_promote_spaces() {
	force_release_enabled && return 0
	anounce "Checking if we can promote spaces"

	status_code=$(http_status_for "https://insights.nyc3.digitaloceanspaces.com/apt/pool/main/main/d/do-agent/do-agent_${VERSION}_amd64.deb")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been promoted. Deploy a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

function promote_spaces() {
	check_can_promote_spaces
	pull_spaces /

	anounce "Promoting packages"
	cp -Luv "$PROJECT_ROOT/repos/apt-beta/pool/main/main/d/do-agent/do-agent_${VERSION}_amd64.deb" "$PROJECT_ROOT/repos/apt/pool/main/main/d/do-agent/"
	cp -Luv "$PROJECT_ROOT/repos/apt-beta/pool/main/main/d/do-agent/do-agent_${VERSION}_i386.deb" "$PROJECT_ROOT/repos/apt/pool/main/main/d/do-agent/"
	cp -Luv "$PROJECT_ROOT/repos/yum-beta/i386/do-agent.${VERSION}.i386.rpm" "$PROJECT_ROOT/repos/yum/i386/"
	cp -Luv "$PROJECT_ROOT/repos/yum-beta/x86_64/do-agent.${VERSION}.amd64.rpm" "$PROJECT_ROOT/repos/yum/x86_64/"

	rebuild_apt_main_packages
	rebuild_yum_main_packages

	# sync the packages first to prevent race conditions
	push_spaces "/apt/pool/main/" --exclude "*" --include "**/*.deb"
	push_spaces "/yum/" --exclude "*" --include "*/*.rpm"

	# then sync the metadata and everything else
	push_spaces "/apt/dists/main/"
	push_spaces "/yum/"
}

function check_can_promote_github() {
	force_release_enabled && return 0
	anounce "Checking the state of Github release $VERSION"
	github_curl \
		--fail \
		-D /dev/stderr \
		"$(github_release_url)" \
		| jq -r '. | select(.prerelease == true) | "Found Release: \(.url)"' \
		| grep . \
		|| abort "Could not find a prerelease version $VERSION to promote. Has it already been released?"
}

function promote_github() {
	check_can_promote_github

	anounce "Removing prerelease flag from '$VERSION' on Github"
	github_curl \
		--fail \
		-i \
		-X PATCH \
		--data-binary '{"prerelease":false}' \
		"$(github_release_url)" \
		| grep 'HTTP/1.1'
}

function check_can_promote_docker() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://hub.docker.com/v2/repositories/digitalocean/do-agent/tags/${VERSION}/")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been released. Deploy a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

function promote_docker() {
	check_can_promote_docker
	anounce "Tagging docker $VERSION-rc as $VERSION"

	docker_login
	local rc="$DOCKER_IMAGE:$VERSION-rc"
	IFS=. read -r major minor _ <<<"$VERSION"

	docker pull "$rc"

	tags="latest $major $major.$minor $VERSION"
	for tag in $tags; do
		docker tag "$rc" "$DOCKER_IMAGE:$tag"
		docker push "$DOCKER_IMAGE:$tag"
	done
}

function http_status_for() {
	url=${1:-}
	[ -z "$url" ] && abort "Usage: ${FUNCNAME[0]} <url>"
	curl -LISsL "$url" | grep 'HTTP/1.1' | awk '{ print $2 }'
}

# print the content type header for the provided file
function content_type_for() {
	file=${1:-}
	[ -z "$file" ] && abort "Usage: ${FUNCNAME[0]} <file>"
	case $file in
		*.deb) echo "application/vnd.debian.binary-package" ;;
		*.rpm) echo "application/x-rpm" ;;
		*.tar.gz) echo "application/gzip" ;;
		*) echo "application/octet-stream"
	esac
}

# get the asset upload URL for VERSION
function github_asset_upload_url() {
	if base=$(github_release_url); then
		echo "${base/api/uploads}/assets"
	else
		return 1
	fi
}

# get the base release url for VERSION
function github_release_url() {
	github_curl \
		--fail \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '.url' \
		| grep .
}


function rm_old_assets() {
	assets=$(github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '.assets[].url')
	for asset in $assets; do
		echo "Removing old asset $asset"
		github_curl \
			-X DELETE \
			"$asset" &
		wait
	done
}

# create a github release for VERSION
function create_github_release() {
	anounce "Checking for existing Github release"
	if github_release_url >/dev/null; then
		echo "Github release exists $VERSION"
		# we cannot upload the same asset twice so we have to delete
		# the old assets before we can commense with uploads
		rm_old_assets || abort "failed to purge Github release assets"
		return 0
	fi

	anounce "Creating Github release $VERSION"

	data=$(cat <<-EOF
	{ "tag_name": "$VERSION", "name": "$VERSION", "prerelease": true, "target_commitish": "master" }
	EOF
	)
	echo "POST: $data"
	github_curl \
		-o /dev/null \
		-X POST \
		-H 'Content-Type: application/json' \
		-d "$data" \
		https://api.github.com/repos/digitalocean/do-agent/releases
}

function docker_login() {
	# gocd has an old version of docker that does not have --pasword-stdin
	docker login -u "$DOCKER_USER" -p "$DOCKER_PASSWORD"
}

function check_can_deploy_docker() {
	force_release_enabled && return 0
	status_code=$(http_status_for "https://hub.docker.com/v2/repositories/digitalocean/do-agent/tags/${VERSION}-rc/")
	case $status_code in
		404)
			return 0
			;;
		200)
			abort "'$VERSION' has already been released. Add a new git tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a version already exists. Try again? Got status code '$status_code'"
			;;
	esac
}

# build and push the RC docker hub image. This image is considered unstable
# and should only be used for testing purposes
function deploy_docker() {
	check_can_deploy_docker
	anounce "Pushing docker images"
	docker_login

	docker build . -t "$DOCKER_IMAGE:unstable"
	tags="${VERSION}-rc"

	for tag in $tags; do
		docker tag "$DOCKER_IMAGE:unstable" "$DOCKER_IMAGE:$tag"
	done

	for tag in $tags unstable; do
		docker push "$DOCKER_IMAGE:$tag"
	done
}

# list the artifacts within the target/ directory
function target_files() {
	find target/pkg  -type f -iname "*${VERSION_REGEX}*" \
		| grep .
}

function check_target_files() {
	target_files > /dev/null || abort "No packages for $VERSION were found in target/.  Did you forget to run make?"
}

# call CURL with github authentication
function github_curl() {
	# if user and token are empty then bash will exit because of unbound vars
	curl -SsL \
		--fail \
		-u "${GITHUB_AUTH_USER}:${GITHUB_AUTH_TOKEN}" \
		"$@"
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
}

# print something to STDOUT with formatting
# Usage: anounce "Some message"
#
# Examples:
#    anounce "Begin execution of something"
#    anounce "All is well"
#
function anounce() {
	msg=${1:-}
	[ -z "$msg" ] && abort "Usage: ${FUNCNAME[0]} <msg>"
	echo ":::::::::::::::::::::::::::::::::::::::::::::::::: $msg ::::::::::::::::::::::::::::::::::::::::::::::::::" > /dev/stderr
}

function is_enabled() {
	v=$(echo "${1:-}" | tr '[:upper:]' '[:lower:]')
	if [[ "${v}" =~ ^y(es)?|t(rue)?|1$ ]]; then
		return 0
	else
		return 1
	fi
}


# send a slack notification or fallback to STDERR
# Usage: notify <success> <msg> [link]
#
# Examples:
#    notify 0 "Deployed to Github failed!"
#    notify "true" "Success!" "https://github.com/"
#
function notify() {
	success=${1:-} msg=${2:-} link=${3:-}

	[ -z "${SLACK_WEBHOOK_URL:-}" ] && return 0

	color="green"
	[[ "$success" =~ ^(false|0|no)$ ]] && color="red"

	payload=$(cat <<-EOF
	{
	  "attachments": [
	    {
	      "fallback": "${msg}",
	      "color": "${color}",
	      "title": "${msg}",
	      "title_link": "${link}",
	      "text": "${msg}",
	      "fields": [
		{
		  "title": "App",
		  "value": "do-agent",
		  "short": true
		},
		{
		  "title": "Version",
		  "value": "${VERSION}",
		  "short": true
		},
		{
		  "title": "User",
		  "value": "${USER}",
		  "short": true
		},
		{
		  "title": "Source",
		  "value": "$(hostname -s)",
		  "short": true
		}
	      ]
	    }
	  ]
	}
	EOF
	)

	curl -sS -X POST \
		--fail \
		--data-binary "$payload" \
		"${SLACK_WEBHOOK_URL}" > /dev/null

	# always pass to prevent pipefailures
	return 0
}

function notify_exit() {
	if [ "$1" != "0" ]; then
		notify 0 "Deploy failed" "${CI_LOG_URL:-}"
	else
		notify 1 "Deploy succeeded"
	fi
}
trap 'notify_exit $?' ERR EXIT INT TERM

main "$@"
