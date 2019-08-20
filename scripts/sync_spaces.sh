#!/usr/bin/env bash
set -ueo pipefail

ME=$(basename "$0")
# rclone sha256 is tag v1.48.0
RCLONE_DOCKER_IMAGE=${RCLONE_DOCKER_IMAGE:-docker.io/digitalocean/rclone@sha256:96a2636d25b0cf5ec6e8c09340f3bea79750463d56378fc66ebbd98df86d8fc3}
SRC_SPACES_BUCKET_NAME="${SRC_SPACES_BUCKET_NAME:-}"
SRC_SPACES_REGION="${SRC_SPACES_REGION:-}"
SRC_SPACES_ENDPOINT="${SRC_SPACES_REGION}.digitaloceanspaces.com"

DEST_SPACES_BUCKET_NAME="${DEST_SPACES_BUCKET_NAME:-}"
DEST_SPACES_REGION="${DEST_SPACES_REGION:-}"
DEST_SPACES_ENDPOINT="${DEST_SPACES_REGION}.digitaloceanspaces.com"

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

	    Execute rclone to sync two DigitalOcean Spaces buckets. Used
	    to create a backup of the primary Linux packages to another region.

	ENVIRONMENT:

	    AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY (required)
	        DigitalOcean Spaces access credentials

	    SRC_SPACES_BUCKET_NAME, SRC_SPACES_REGION (required)
	        DigitalOcean Spaces bucket information for the source bucket.
	        This bucket will be used as the source bucket.
	
	    DEST_SPACES_BUCKET_NAME, DEST_SPACES_REGION (required)
	        DigitalOcean Spaces bucket information for the destination bucket.
	        This bucket will be used as the destination bucket. All files
	        Will be copied from SRC_SPACES_* to DEST_SPACES_*
	
	    SLACK_WEBHOOK_URL (optional)
	        Webhook URL to send notifications. Enables Slack
	        notifications

	COMMANDS:

	    sync
	        Sync everything from SRC_SPACES_* to DEST_SPACES_*

	EOF
}

function main() {
	cmd=${1:-}

	case "$cmd" in
		sync)
			sync_spaces
			;;
		help|--help|-h)
			usage
			exit 0
			;;
		*)
			abort "Unknown command '$cmd'. See $ME --help for available commands"
			;;
	esac
}

function check_can_sync_spaces() {
	announce "Checking configuration"
	[ -z "${SRC_SPACES_BUCKET_NAME}" ] && abort "SRC_SPACES_BUCKET_NAME is not set"
	[ -z "${SRC_SPACES_REGION}" ] && abort "SRC_SPACES_REGION is not set"
	[ -z "${DEST_SPACES_BUCKET_NAME}" ] && abort "DEST_SPACES_BUCKET_NAME is not set"
	[ -z "${DEST_SPACES_REGION}" ] && abort "DEST_SPACES_REGION is not set"
	[ -z "${AWS_SECRET_ACCESS_KEY}" ] && abort "AWS_SECRET_ACCESS_KEY is not set"
	[ -z "${AWS_ACCESS_KEY_ID}" ] && abort "AWS_ACCESS_KEY_ID is not set"

	return 0
}

function sync_spaces() {
	check_can_sync_spaces

	announce "Creating rclone.conf"
	cat <<-EOF > rclone.conf
	[source]
	type = s3
	provider = DigitalOcean
	env_auth = true
	region =
	endpoint = ${SRC_SPACES_ENDPOINT}
	acl = public-read

	[dest]
	type = s3
	provider = DigitalOcean
	env_auth = true
	region =
	endpoint = ${DEST_SPACES_ENDPOINT}
	acl = public-read
	EOF

	announce "Pulling ${RCLONE_DOCKER_IMAGE}"
	quiet_docker_pull "${RCLONE_DOCKER_IMAGE}"

	announce "Syncing spaces buckets"
	docker run \
		-e "AWS_SECRET_ACCESS_KEY" \
		-e "AWS_ACCESS_KEY_ID" \
		--rm -i \
		--net=host \
		-v="${PWD}/rclone.conf:/root/.config/rclone/rclone.conf" \
		"${RCLONE_DOCKER_IMAGE}" \
		sync "source:${SRC_SPACES_BUCKET_NAME}" "dest:${DEST_SPACES_BUCKET_NAME}" \
		--checksum \
		-v \
		--delete-during

	announce "Sync spaces complete"
}

function quiet_docker_pull() {
	img=${1:-}
	[ -z "$img" ] && abort "img param is required. Usage: ${FUNCNAME[0]} <img>"
	docker pull "${img}" | grep -e 'Pulling from' -e Digest -e Status -e Error
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ABORT ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
}

# error with an error message
function error() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
}

# print something to STDOUT with formatting
# Usage: announce "Some message"
#
# Examples:
#    announce "Begin execution of something"
#    announce "All is well"
#
function announce() {
	msg=${1:-}
	[ -z "$msg" ] && abort "Usage: ${FUNCNAME[0]} <msg>"
	echo ":::::::::::::::::::::::::::::::::::::::::::::::::: $msg ::::::::::::::::::::::::::::::::::::::::::::::::::" > /dev/stderr
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
	color="green"
	if [[ "$success" =~ ^(false|0|no)$ ]]; then
		color="red"
	fi

	if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
		echo "${msg}"
		return 0
	fi

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
		  "value": "do-agent-repo-sync",
		  "short": true
		},
		{
		  "title": "User",
		  "value": "$(whoami)@$(hostname -s)",
		  "short": true
		}
	      ]
	    }
	  ]
	}
	EOF
	)

	curl -sSL -X POST \
		--fail \
		--data-binary "$payload" \
		"${SLACK_WEBHOOK_URL}" > /dev/null || true

	# always pass to prevent pipefailures
	return 0
}

function notify_exit() {
	if [ "${1:0}" != "0" ]; then
		notify 0 "Sync spaces buckets failed" "${CI_LOG_URL:-}" || true
	else
		notify 1 "Sync spaces buckets succeeded" || true
	fi
}
trap 'notify_exit $?; rm -rf rclone.conf' ERR EXIT INT TERM

main "$@"
