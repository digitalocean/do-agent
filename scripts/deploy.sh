#!/usr/bin/env bash
set -ueo pipefail
#set -x

ME=$(basename "$0")
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
GPG_PRIVATE_KEY="$PROJECT_ROOT/sonar-agent.key"
DOCKER_IMAGE="docker.io/digitalocean/do-agent"
VERSION=${VERSION:-$(cat target/VERSION || true)}

# CI_LOG_URL=""
# if [ -n "${GO_SERVER_URL:-}" ]; then
# 	CI_LOG_URL=${GO_SERVER_URL}/tab/${GO_STAGE_NAME}/detail/${GO_PIPELINE_NAME}/${GO_PIPELINE_COUNTER}/${GO_STAGE_NAME}/${GO_STAGE_COUNTER}/${GO_JOB_NAME}
# fi

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
			check_released
			deploy_github
			;;
		spaces)
			check_version
			check_target_files
			check_released
			deploy_spaces
			;;
		docker)
			check_version
			check_target_files
			check_released
			deploy_docker
			;;
		promote)
			check_version
			check_released
			promote_github
			promote_spaces
			promote_docker
			;;
		all)
			check_version
			check_target_files
			check_released
			deploy_spaces
			deploy_github
			deploy_docker
			;;
		help)
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

# if a release with the VERSION tag is already published then we cannot deploy
# this version over the previous release.
function check_released() {
	if [[ "${FORCE_RELEASE:-}" =~ ^(y|yes|1|true)$ ]]; then
		echo
		echo "WARNING! forcing a release of $VERSION"
		echo
		return 0
	fi

	status=$(curl -LI -SsL "https://insights.nyc3.digitaloceanspaces.com/yum-beta/x86_64/do-agent.${VERSION}.amd64.rpm" | grep 'HTTP/1.1')

	case "$status" in
		*'HTTP/1.1 404 Not Found'*)
			return 0
			;;
		*'HTTP/1.1 200 OK'*)
			abort "'$VERSION' has already been released. So it this tag can no longer be deployed. To deploy again you must add a new tag or use pass FORCE_RELEASE=1."
			;;
		*)
			abort "Failed to check if a stable version already exists. Try again? got -> $status"
			;;
	esac
}

function verify_gpg_key() {
	stat "$GPG_PRIVATE_KEY" > /dev/null || abort "$GPG_PRIVATE_KEY is required"
}

function deploy_spaces() {
	pull_spaces

	anounce "Copying apt and yum packages"

	target_files | grep -P '\.deb$' | while IFS= read -r file; do
		cp -Luv "$file" repos/apt/pool/beta/main/d/do-agent/
	done

	target_files | grep -P '\.rpm$' | while IFS= read -r file; do
		dest=repos/yum-beta/x86_64/
		[[ "$file" =~ "i386" ]] && \
			dest=repos/yum-beta/i386/
		cp -Luv "$file" "$dest"
	done

	rebuild_apt_packages
	rebuild_yum_packages

	push_spaces
}

function rebuild_apt_packages() {
	verify_gpg_key
	anounce "Rebuilding apt package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/apt:/work/apt" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-apt" \
		|| abort "Failed to rebuild apt package indexes"
}

function rebuild_yum_packages() {
	verify_gpg_key
	anounce "Rebuilding yum package indexes"
	docker run \
		--rm -i \
		--net=host \
		-v "${PROJECT_ROOT}/repos/yum:/work/yum" \
		-v "${PROJECT_ROOT}/repos/yum-beta:/work/yum-beta" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-yum" \
		|| abort "Failed to rebuild yum package indexes"
}

function pull_spaces() {
	anounce "Pulling remote packages"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		s3://insights/ \
		./repos/ \
		--quiet \
		--delete \
		--acl public-read
}

function push_spaces() {
	anounce "Pushing package changes"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		./repos/ \
		s3://insights/ \
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


# deploy the compiled binaries and packages to github releases
function deploy_github() {
	if ! create_github_release ; then
		abort "Github deploy failed"
	fi

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

function promote_spaces() {
	pull_spaces

	anounce "Copying deb and rpm packages to main channels"
	cp -Luv "$PROJECT_ROOT/repos/apt/pool/beta/main/d/do-agent/do-agent_${VERSION}_.deb" "$PROJECT_ROOT/repos/apt/pool/main/main/d/do-agent/"
	cp -Luv "$PROJECT_ROOT/repos/yum-beta/i386/do-agent.${VERSION}.i386.rpm" "$PROJECT_ROOT/repos/yum/i386/"
	cp -Luv "$PROJECT_ROOT/repos/yum-beta/x86_64/do-agent.${VERSION}.amd64.rpm" "$PROJECT_ROOT/repos/yum/x86_64/"

	rebuild_apt_packages
	rebuild_yum_packages

	push_spaces
}

function promote_github() {
	anounce "Removing prerelease flag from github release"

	github_curl \
		-D /dev/stderr \
		-X PATCH \
		--data-binary '{"prerelease":false}' \
		"$(github_release_url)"
}

function promote_docker() {
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
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '. | "https://api.github.com/repos/digitalocean/do-agent/releases/\(.id)"' \
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
	if github_release_url; then
		echo "Github release exists $VERSION"
		# we cannot upload the same asset twice so we have to delete
		# the old assets before we can commense with uploads
		rm_old_assets || abort "failed to purge Github release assets"
		return 0
	fi

	echo "Creating Github release $VERSION"

	data=$(cat <<-EOF
	{ "tag_name": "$VERSION", "prerelease": true, "target_commitish": "beta" }
	EOF
	)
	echo "$data"
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

# build and push the RC docker hub image. This image is considered unstable
# and should only be used for testing purposes
function deploy_docker() {
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
	find target/pkg -type f -iname "do-agent[._]${VERSION}[._]*" | grep .
}

function check_target_files() {
	target_files || abort "No packages for $VERSION were found in target/.  Did you forget to run make?"
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


# send a slack notification or fallback to STDERR
# Usage: notify_slack <success> <msg> [link]
#
# Examples:
#    notify 0 "Deployed to Github failed!"
#    notify "true" "Success!" "https://github.com/"
#
function notify() {
	success=${1:-}
	msg=${2:-}
	link=${3:-}

	if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
		{
			echo
			echo "SLACK_WEBHOOK_URL is unset. Falling back to stderr"
			echo "(success:$success) $msg $link"
			echo
		} > /dev/stderr
		return 0
	fi

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


main "$@"
