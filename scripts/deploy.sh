#!/usr/bin/env bash
set -ueo pipefail
#set -x

ME=$(basename "$0")
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
VERSION="${VERSION:-$(cat target/VERSION 2>/dev/null)}"
PKG_VERSION="$(git describe --tags | grep -oP '[0-9]+\.[0-9]+\.[0-9]+(-[0-9]+)?')"
GPG_PRIVATE_KEY="$PROJECT_ROOT/sonar-agent.key"
DOCKER_IMAGE="docker.io/digitalocean/do-agent"

CI_LOG_URL=""
if [ -n "${GO_SERVER_URL:-}" ]; then
	CI_LOG_URL=${GO_SERVER_URL}/tab/${GO_STAGE_NAME}/detail/${GO_PIPELINE_NAME}/${GO_PIPELINE_COUNTER}/${GO_STAGE_NAME}/${GO_STAGE_COUNTER}/${GO_JOB_NAME}
fi

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
			check_stable
			deploy_github
			;;
		spaces)
			check_version
			check_target_files
			check_stable
			deploy_spaces
			;;
		docker)
			check_version
			check_target_files
			check_stable
			deploy_docker
			;;
		promote)
			check_version
			check_stable
			promote_github
			promote_spaces
			promote_docker
			;;
		all)
			check_version
			check_target_files
			check_stable
			deploy_github
			deploy_spaces
			deploy_docker
			;;
		help)
			usage
			exit 0
			;;
		*)
			{
				echo
				echo "Unknown command '$cmd'"
				echo
				usage
			} > /dev/stderr
			exit 1
			;;
	esac
}

# verify the VERSION env var
function check_version() {
	[[ "${VERSION}" =~ v[0-9]+\.[0-9]+\.[0-9]+ ]] || \
		abort "VERSION env var should be semver format (e.g. v0.0.1)"
}

# if a release with the VERSION tag is already published without the prerelease
# flag set to true then we cannot deploy or we risk overriding binaries that
# have already been considered stable. At this point a new tag with an
# incremented patch version is required. 
#
# This is not considered an error. It just means there is no additional deploy
# necessary
function check_stable() {
	stable_release=$(github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases" \
		| jq -r ".[] | select(.tag_name == \"$VERSION\") | select(.prerelease == false) | .id")
	if [ -n "$stable_release" ]; then
		{
			echo
			echo "Github release '$VERSION' has already been released to main. So it this tag can no longer be deployed."
			echo "To deploy again you must add a new tag."
			echo
		} > /dev/stderr
		exit 0
	fi
}

function verify_gpg_key() {
	stat "$GPG_PRIVATE_KEY" > /dev/null || abort "$GPG_PRIVATE_KEY is required"
}

function deploy_spaces() {
	pull_spaces

	anounce "Moving built deb packages"
	for file in $(target_files | grep -P '\.deb$'); do
		cp "$file" repos/apt/pool/beta/main/d/do-agent/
	done

	anounce "Moving yum packages"
	for file in $(target_files | grep -P '\.rpm$'); do
		dest=repos/yum-beta/x86_64/
		[[ "$file" =~ "i386" ]] && \
			dest=repos/yum-beta/i386/
		cp "$file" "$dest"
	done

	rebuild_apt_packages

	rebuild_yum_packages

}

function rebuild_apt_packages() {
	verify_gpg_key
	anounce "Rebuilding apt package indexes"
	docker run \
		--rm \
		--net=host \
		-v "${PROJECT_ROOT}/repos/apt:/work/apt" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		-ti \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-apt" || exit 1
}

function rebuild_yum_packages() {
	verify_gpg_key
	anounce "Rebuilding yum package indexes"
	docker run \
		--rm \
		--net=host \
		-v "${PROJECT_ROOT}/repos/yum:/work/yum" \
		-v "${PROJECT_ROOT}/repos/yum-beta:/work/yum-beta" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		-ti \
		"docker.internal.digitalocean.com/eng-insights/agent-packager-yum" || exit 1
}

function pull_spaces() {
	anounce "Pulling remote packages"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		s3://insights/ \
		./repos/ \
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
		--rm -t "$(tty &>/dev/null && echo '-i')" \
		-e "AWS_ACCESS_KEY_ID=${SPACES_ACCESS_KEY_ID}" \
		-e "AWS_SECRET_ACCESS_KEY=${SPACES_SECRET_ACCESS_KEY}" \
		-e "AWS_DEFAULT_REGION=nyc3" \
		-v "$(pwd):/project" \
		-w /project \
		-u "$(id -u)" \
		mesosphere/aws-cli \
		"$@"
}


# deploy the compiled binaries and packages to github releases
function deploy_github() {
	if ! create_github_release ; then
		echo "Aborting github deploy"
		exit 1
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

	anounce "Copying deb packages to main channel"
	cp "$PROJECT_ROOT/repos/apt/pool/beta/main/d/do-agent/do-agent_${PKG_VERSION}_.deb" "$PROJECT_ROOT/repos/apt/pool/main/main/d/do-agent/"

	anounce "Copying yum packages to main channel"
	cp "$PROJECT_ROOT/repos/yum-beta/i386/do-agent.${PKG_VERSION}.i386.rpm" "$PROJECT_ROOT/repos/yum/i386/"
	cp "$PROJECT_ROOT/repos/yum-beta/x86_64/do-agent.${PKG_VERSION}.amd64.rpm" "$PROJECT_ROOT/repos/yum/x86_64/"

	rebuild_apt_packages
	rebuild_yum_packages

	push_spaces
}

function promote_github() {
	github_curl \
		-D /dev/stderr \
		-X PATCH \
		--data-binary '{"prerelease":false}' \
		"$(github_release_url)"
}

function promote_docker() {
	docker_login
	local version=${VERSION/v}
	local rc="$DOCKER_IMAGE:$version-rc"
	IFS=. read -r major minor _ <<<"$version"

	docker pull "$rc"

	tags="latest $major $major.$minor $version"
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
		| jq -r '. | select(.prerelease == true) | "https://api.github.com/repos/digitalocean/do-agent/releases/\(.id)"'
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
	if [ -n "$(github_release_url)" ]; then
		echo "Github release exists $VERSION"
		# we cannot upload the same asset twice so we have to delete
		# the old assets before we can commense with uploads
		rm_old_assets
	else
		echo "Creating Github release $VERSION"
		data="{ \"tag_name\": \"$VERSION\", \"prerelease\": true }"
		echo "$data"

		github_curl \
			-o /dev/null \
			-X POST \
			-H 'Content-Type: application/json' \
			-d "$data" \
			https://api.github.com/repos/digitalocean/do-agent/releases
	fi
}

function docker_login() {
	docker login -u "$DOCKER_USER" --password-stdin <<<"$DOCKER_PASSWORD"  
}

# build and push the RC docker hub image. This image is considered unstable
# and should only be used for testing purposes
function deploy_docker() {
	docker_login

	docker build . -t "$DOCKER_IMAGE:unstable"
	tags="${VERSION/v}-rc"

	for tag in $tags; do
		docker tag "$DOCKER_IMAGE:unstable" "$DOCKER_IMAGE:$tag"
	done

	for tag in $tags unstable; do
		docker push "$DOCKER_IMAGE:$tag"
	done
}

# list the artifacts within the target/ directory
function target_files() {
	find target/pkg -type f -iname "*${PKG_VERSION}*" || \
		abort "No packages for $PKG_VERSION were found in target/.  Did you forget to run make?"
}

function check_target_files() {
	target_files
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


function cp() {
	src=${1:-}
	dest=${2:-}
	cp -Luv "$src" "$dest" || exit 1
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
		  "value": "${PKG_VERSION}",
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
