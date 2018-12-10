#!/usr/bin/env bash
set -ueo pipefail
# set -x

# UBUNTU_VERSIONS="trusty utopic vivid wily xenial yakkety zesty artful bionic cosmic"
# DEBIAN_VERSIONS="wheezy jessie stretch buster"
# RHEL_VERSIONS="6 7"
# FEDORA_VERSIONS="27 28"

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
VERSION="${VERSION:-$(cat target/VERSION 2>/dev/null)}"

# display usage for this script
function usage() {
	cat <<-EOF
	NAME:
	
	  $(basename "$0")
	
	SYNOPSIS:

	  $0 <cmd>

	DESCRIPTION:

	    The purpose of this script is to publish build artifacts
	    to Github and docker.io. RPM and DEB files are packaged
	    and deployed to apt/yum with DigitalOcean's internal
	    build system after they are considered stable.

	ENVIRONMENT:
		
	    VERSION (required)
	        The version to publish

	    GITHUB_AUTH_USER (required)
	        Github user to use for publishing to Github

	    GITHUB_AUTH_TOKEN (required)
	        Github access token to use for publishing to Github

	    DOCKER_USER (required)
	        hub.docker.io username to use for pushing builds

	    DOCKER_PASSWORD (required)
	        hub.docker.io password to use for pushing builds

	    SPACES_ACCESS_KEY_ID (required)
	        Spaces key ID to use for DO Spaces deployment

	    SPACES_SECRET_ACCESS_KEY (required)
	        Spaces secret access key ID to use for DO Spaces deployment

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
			if deploy_github ; then
				notify_slack "true" "Deployed packages to Github!" "https://github.com/digitalocean/do-agent/releases/tag/$VERSION"
			else
				notify_slack "false" "Failed to deploy packages to Github!" "${TRAVIS_BUILD_WEB_URL:-}"
			fi
			;;
		spaces)
			check_version
			if deploy_spaces; then
				notify_slack "true" "Deployed artifacts to Spaces"
			else
				notify_slack "false" "Failed to deploy artifacts to Spaces" "${TRAVIS_BUILD_WEB_URL:-}"
			fi
			;;
		docker)
			check_version
			if deploy_dockerhub ; then
				notify_slack "true" "Deployed images to docker" "https://hub.docker.com/r/digitalocean/do-agent"
			else
				notify_slack "false" "failed to deploy images to docker" "${TRAVIS_BUILD_WEB_URL:-}"
			fi
			;;
		promote)
			check_version
			promote_packages
			promote_github
			promote_docker
			;;
		all)
			check_version
			main "github"
			main "spaces"
			main "docker"
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

function deploy_spaces() {
	# this file is required to sign the packages
	GPG_PRIVATE_KEY="$PROJECT_ROOT/sonar-agent.key"
	stat "$GPG_PRIVATE_KEY"

	anounce "Pulling remote packages"
	aws s3 \
		--endpoint-url https://nyc3.digitaloceanspaces.com \
		sync \
		s3://insights/ \
		./repos/ \
		--acl public-read

	anounce "Moving built packages to repos/apt"
	files=$(target_files | grep -P '.deb$')
	for file in $files; do
		cp -Huv "$file" repos/apt/pool/beta/main/d/do-agent/
	done

	anounce "Moving build packages to repos/yum-beta"
	files=$(target_files | grep -P '.rpm$' | grep amd64)
	for file in $files; do
		cp -Huv "$file" repos/yum-beta/x86_64/
	done
	files=$(target_files | grep -P '.rpm$' | grep i386)
	for file in $files; do
		cp -Huv "$file" repos/yum-beta/i386/
	done

	anounce "Rebuilding apt package indexes"
	docker run \
		--rm \
		--net=host \
		-v "${PROJECT_ROOT}/repos/apt:/work/apt" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		-ti \
		"docker.io/digitalocean/agent-packager-apt"

	anounce "Rebuilding yum package indexes"
	docker run \
		--rm \
		--net=host \
		-v "${PROJECT_ROOT}/repos/yum:/work/yum" \
		-v "${PROJECT_ROOT}/repos/yum-beta:/work/yum-beta" \
		-v "${PROJECT_ROOT}/sonar-agent.key:/work/sonar-agent.key:ro" \
		-w /work \
		-ti \
		"docker.io/digitalocean/agent-packager-yum"

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
	# if a release with this tag is already published without the prerelease
	# flag set to true then we cannot deploy or we risk overriding binaries
	# that have already been considered stable. At this point a new tag with
	# an incremented patch version is required
	stable_release=$(github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases" \
		| jq -r ".[] | select(.tag_name == \"$VERSION\") | select(.prerelease == false) | .id")
	if [ -n "$stable_release" ]; then
		{
			echo "Github release '$VERSION' is not flagged as prerelease. Refusing to deploy."
			echo "To deploy a fix to a stable release you must increment the fix version with a new tag."
		} > /dev/stderr
		exit 0
	fi


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

# build and push the RC docker hub image. This image is considered unstable
# and should only be used for testing purposes
function deploy_dockerhub() {
	docker login -u "$DOCKER_USER" --password-stdin <<<"$DOCKER_PASSWORD"  

	image="docker.io/digitalocean/do-agent"
	docker build . -t "$image:unstable"
	tags="${VERSION/v}-rc"

	for tag in $tags; do
		docker tag "$image:unstable" "$image:$tag"
	done

	for tag in $tags unstable; do
		docker push "$image:$tag"
	done
}

# list the artifacts within the target/ directory
function target_files() {
	v=${VERSION/v}
	if ! packages=$(find target/pkg -type f -iname "*$v*"); then
		abort "No packages for $VERSION were found in target/.  Did you forget to run make?"
	fi

	ls target/do-agent_linux_*
	echo "$packages"
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



# send a slack notification
# Usage: notify_slack <success> <msg> [link]
#
# Examples:
#    notify_slack 0 "Deployed to Github failed!"
#    notify_slack "true" "Success!" "https://github.com/"
#
function notify_slack() {
	if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
		echo "SLACK_WEBHOOK_URL is unset. Not sending notification" > /dev/stderr
		return 0
	fi

	success=${1:-}
	msg=${2:-}
	link=${3:-}

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
		--data "$payload" \
		"${SLACK_WEBHOOK_URL}" > /dev/null

	# always pass to prevent pipefailures
	return 0
}


main "$@"
