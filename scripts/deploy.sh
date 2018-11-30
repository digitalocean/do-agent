#!/usr/bin/env bash
set -ueo pipefail
# set -x

# UBUNTU_VERSIONS="trusty utopic vivid wily xenial yakkety zesty artful bionic cosmic"
# DEBIAN_VERSIONS="wheezy jessie stretch buster"
# RHEL_VERSIONS="6 7"
# FEDORA_VERSIONS="27 28"

VERSION="${VERSION:-$(cat target/VERSION 2>/dev/null)}"

# display usage for this script
function usage() {
	cat <<-EOF
	NAME:
	
	  $(basename "$0")
	
	SYNOPSIS:

	  $0 <cmd>

	DESCRIPTION:

	    The purpose of this script is to publish build artifacts.
	    Deployments push artifacts in Prerelease or BETA mode. After they
	    are published and tested in the BETA/Prerelease phase they can then
	    be promoted.

	ENVIRONMENT:
		
	    VERSION (required)
	         The version to publish or promote

	    GITHUB_AUTH_USER
	         Github user to use for publishing to Github
	         Required for github deploy

	    GITHUB_AUTH_TOKEN
	         Github access token to use for publishing to Github
	         Required for github deploy

	    DOCKER_USER
	         User used to login to docker.io
	         Required for docker deploy

	    DOCKER_PASSWORD
	         Password used to login to docker.io
	         Required for docker deploy

	    SLACK_WEBHOOK_URL
	         Webhook URL to send notifications
	         Optional: enables Slack notifications

	COMMANDS:
	
	    github
	         push target/ assets to github

	    dockerhub
	         build and push docker containers to public docker hub

	    all
	         push to github and dockerhub

	    promote
	         promote VERSION from beta to upstream and remove
	         the prerelease flag from the Github release

	EOF
}

function main() {
	cmd=${1:-}

	case "$cmd" in
		all)
			check_version
			deploy_github
			deploy_dockerhub
			wait
			;;
		github)
			check_version
			deploy_github
			;;
		promote)
			check_version
			promote_github
			;;
		dockerhub)
			check_version
			deploy_dockerhub
			;;
		help)
			usage
			exit 0
			;;
		*)
			usage
			exit 1
			;;
	esac
}

# verify the VERSION env var
function check_version() {
	[[ "${VERSION}" =~ v[0-9]+\.[0-9]+\.[0-9]+ ]] || \
		abort "VERSION env var should be semver format (e.g. v0.0.1)"
}

# build and push docker images
function deploy_dockerhub() {
	echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USER" --password-stdin

	version=${VERSION/v}
	IFS=. read -r major minor _ <<<"$version"

	image="docker.io/digitalocean/do-agent"
	docker build . -t "$image:$version"
	tags="latest $major $major.$minor"

	for tag in $tags; do
		docker tag "$image:$version" "$image:$tag"
	done

	for tag in $tags $version; do
		docker push "$image:$tag"
	done
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
			-H 'Content-Type: application/octet-stream' \
			-d "@${file}" \
			"$upload_url?name=$name" \
			| jq -r '. | "Success: \(.name)"' &
	done
	wait
}

# remove the prerelease flag from the github release for VERSION
function promote_github() {
	if ! url=$(github_release_url); then
		abort "Github release for $VERSION does not exist"
	fi

	echo "Removing github prerelease tag for $VERSION"
	github_curl \
		-o /dev/null \
		-X PATCH \
		-H 'Content-Type: application/json' \
		-d '{ "prerelease": false }' \
		"$url"
}

# verify the deb version exists in the beta repository before attempting to promote
function check_deb_version() {
	v=${VERSION/v}

	abort "Unable to check deb version because the new URL has not been set"

	url="https://????????????????-beta/packages/ubuntu/zesty/do-agent_${v}_amd64.deb"
	echo "Checking for version $v"
	curl --fail \
		-SsLI \
		"${url}" \
		| grep 'HTTP/1'
}

# get the asset upload URL for VERSION
function github_asset_upload_url() {
	github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '. | "https://uploads.github.com/repos/digitalocean/do-agent/releases/\(.id)/assets"'
}

# get the base release url for VERSION
function github_release_url() {
	github_curl \
		"https://api.github.com/repos/digitalocean/do-agent/releases/tags/$VERSION" \
		| jq -r '. | "https://api.github.com/repos/digitalocean/do-agent/releases/\(.id)"'
}


# create a github release for VERSION
function create_github_release() {
	if github_asset_upload_url; then
		echo "Github release exists $VERSION"
		return 0
	fi

	echo "Creating Github release $VERSION"
	data="{ \"tag_name\": \"$VERSION\", \"prerelease\": true }"
	echo "$data"

	github_curl \
		-o /dev/null \
		-X POST \
		-H 'Content-Type: application/json' \
		-d "$data" \
		https://api.github.com/repos/digitalocean/do-agent/releases
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

# ask the user for input
function ask() {
	question=${1:-}
	[ -z "$question" ] && abort "Usage: ${FUNCNAME[0]} <question>"
	read -p "$question " -r
	echo -n "$REPLY"
}

# ask the user for a yes or no answer. Returns 0 for yes or 1 for no.
function confirm() {
	question="$1 (y/n)"
	yn=$(ask "$question")
	echo
	[[ $yn =~ ^[Yy]$ ]] && return 0
	return 1
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
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
		echo "env var SLACK_WEBHOOK_URL is unset. Not sending notification" > /dev/stderr
		return 0
	fi

	success=${1:-}
	msg=${3:-}
	link=${2:-}

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
	      "fields": [
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
