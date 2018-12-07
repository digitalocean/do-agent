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

	    The purpose of this script is to publish build artifacts
	    to Github. They are then packaged and deployed to
	    apt/yum/docker with DigitalOcean's internal build system.

	ENVIRONMENT:
		
	    VERSION (required)
	        The version to publish

	    GITHUB_AUTH_USER
	         Github user to use for publishing to Github
	         Required for github deploy

	    GITHUB_AUTH_TOKEN
	         Github access token to use for publishing to Github
	         Required for github deploy

	    SLACK_WEBHOOK_URL
	         Webhook URL to send notifications
	         Optional: enables Slack notifications

	COMMANDS:
	
	    github
	         push target/ assets to github

	EOF
}

function main() {
	cmd=${1:-}

	case "$cmd" in
		github)
			check_version
			deploy_github
			notify_slack "true" "Deployed packages to Github!" "https://github.com/digitalocean/do-agent/releases/tags/$VERSION"
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
			-H 'Content-Type: application/octet-stream' \
			-d "@${file}" \
			"$upload_url?name=$name" \
			| jq -r '. | "Success: \(.name)"' &
	done
	wait
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
