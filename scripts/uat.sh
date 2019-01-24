#!/usr/bin/env bash
#
# Author: Brett Jones <blockloop>
# Purpose: Provide simple UAT tasks for creating/updating/deleting droplets
# configured with do-agent. To add another method to this script simply
# create a new function called 'function command_<task>'. It will automatically
# get picked up as a new command.

set -ue

# optional parameters
# disable installation of the agent
NO_INSTALL=${NO_INSTALL:-0}

# team context in the URL of the browser
CONTEXT=14661f
OS=$(uname | tr '[:upper:]' '[:lower:]')
TAG=do-agent-uat-${USER}
SUPPORTED_IMAGES="centos-6-x32 centos-6-x64 centos-7-x64 debian-8-x32 debian-8-x64 \
	debian-9-x64 fedora-27-x64 fedora-28-x64 ubuntu-14-04-x32 ubuntu-14-04-x64 \
	ubuntu-16-04-x32 ubuntu-16-04-x64 ubuntu-18-04-x64 ubuntu-18-10-x64"

JONES_SSH_FINGERPRINT="a1:bc:00:38:56:1f:d2:b1:8e:0d:4f:9c:f0:dd:66:6d"
THOR_SSH_FINGERPRINT="c6:c6:01:e8:71:0a:58:02:2c:b3:e5:95:0e:b1:46:06"
EVAN_SSH_FINGERPRINT="b9:40:22:bd:fb:d8:fa:fa:4e:11:d9:8e:58:e9:41:73"
SNYDER_SSH_FINGERPRINT="47:31:9b:8b:87:a7:2d:26:79:17:87:83:53:65:d4:b4"

# disabling literal '\n' error in shellcheck since that is the expected
# behavior because it will be added to the JSON request body and
# executed/expanded on the server
# shellcheck disable=SC1117
USER_DATA_DEB="#!/bin/bash \n\
[ -z \`command -v curl\` ] && apt-get -qq update && apt-get install -q -y curl \n\
curl -sL https://insights.nyc3.cdn.digitaloceanspaces.com/install.sh | sudo bash"

# shellcheck disable=SC1117
USER_DATA_RPM="#!/bin/bash \n\
[ -z \`command -v curl\` ] && yum -y install curl \n\
curl -sL https://insights.nyc3.cdn.digitaloceanspaces.com/install.sh | sudo bash"


function main() {
	cmd=${1:-}
	[ -z "$cmd" ] && usage && exit 0
	shift
	fn=command_$cmd
	# disable requirement to quote 'fn' which would break this code
	# shellcheck disable=SC2086
	if [ "$(type -t ${fn})" = function ]; then
		${fn} "$@"
	else
		usage
		exit 1
	fi
}

# show the help text
function command_help() {
	usage
}

# show script usage. This parses every function named command_ from this script
# and displays it as a possible command
function usage() {
	echo
	echo "Usage: $0 [command]"
	echo
	echo "Possible commands: "
	grep -P '^function command_' "$0" \
		| sed 's,function command_,,g' \
		| sed 's,() {,,g' \
		| sort \
		| xargs -n1 -I{} echo "  {}"
	echo
}

# delete all droplets tagged with $TAG
function command_delete() {
	confirm "Are you sure you want to delete all droplets with the tag ${TAG}?" \
		|| (echo "Aborted" && return 1)

	echo "Deleting..."
	request DELETE "/droplets?tag_name=$TAG" \
		| jq .
}

# list all droplet IP addresses tagged with $TAG
function command_list_ips() {
	list_ips
}

# list all droplet IDs tagged with $TAG
function command_list_ids() {
	list_ids
}

# list all droplets with all of their formatted metadata
function command_list() {
	list | jq .
}

# open the browser to show the list of droplets associated with $TAG
function command_browse() {
	launch "https://cloud.digitalocean.com/tags/$TAG?i=${CONTEXT}"
}

# graphs all droplets in the browser
function command_graphs() {
	urls=$(command_list_ids | xargs -n1 -I{} echo https://cloud.digitalocean.com/droplets/{}/graphs?i=${CONTEXT} | tee /dev/stderr)
	if confirm "Open these urls?"; then
		for u in $urls; do
			launch "$u"
		done
	else
		echo "Aborting"
	fi
}

# create a droplet for every SUPPORTED_IMAGE and automatically install do-agent
# using either apt or yum
function command_create() {
	for i in $SUPPORTED_IMAGES; do
		create_image "$i" &
	done
	wait

	if confirm "Open the tag list page?"; then
		launch "https://cloud.digitalocean.com/tags/$TAG?i=${CONTEXT}"
	fi
}

# ssh to all droplets and run <init system> status do-agent to verify
# that it is indeed running
function command_status() {
	command_exec "if command -v systemctl 2&>/dev/null; then \
		systemctl is-active do-agent; \
	else \
		initctl status do-agent; \
	fi"
}

# ssh to all droplets and run yum/apt update to upgrade to the latest published
# version of do-agent
function command_update() {
	exec_rpm "yum -q -y update do-agent"
	exec_deb "apt-get -qq update"
	exec_deb "apt-get -qq install --only-upgrade do-agent"
}

# ssh to all droplets and execute a command
function command_exec() {
	[ -z "$*" ] && abort "Usage: $0 exec <command>"
	exec_ips "$(list_ips)" "$*"
}

# ssh to all debian-based droplets (ubuntu/debian) and execute a command
function command_exec_deb() {
	exec_deb "$*"
}

# ssh to all rpm-based droplets (centos/fedora) and execute a command
function command_exec_rpm() {
	exec_rpm "$*"
}

# list droplet IP addresses for deb based distros
function command_list_ips_deb() {
	list_ips_deb
}

# list droplet IP addresses for rpm based distros
function command_list_ips_rpm() {
	list_ips_rpm
}

# execute a command against a list of IP addresses
# Usage:   exec_ips <ips> <command>
# Example: exec_ips "$(list_ips_rpm)" "yum update DO agent"
function exec_ips() {
	{ [ -z "${1:-}" ] || [ -z "${2:-}" ]; } \
		&& abort "Usage: ${FUNCNAME[0]} <ips> <command>"

	ips=$1
	shift
	script="hostname -s; { $*; }"
	echo "================================================"
	echo "> $*"
	echo "================================================"
	for ip in $ips; do
		# shellcheck disable=SC2029
		echo "$(echo
			echo -n "$(tput setaf 2)>>>> $ip: $(tput sgr 0)"
			ssh -o "StrictHostKeyChecking no" \
				-o "UserKnownHostsFile=/dev/null" \
				-o "LogLevel=ERROR" \
				"root@${ip}" "${script}" 2>/dev/stdout || true
		)" &
	done
	wait
}

# ssh to each droplet one after another. After each connection you will be
# connected to the next in the list unless you press CTRL-C
function command_ssh() {
	for ip in $(list_ips); do
		echo -n ">>>> $ip: "
		# shellcheck disable=SC2029
		ssh -o "StrictHostKeyChecking no" "root@${ip}"
		sleep 0.2
	done
}

# show version information about remote installed versions
function command_versions() {
	exec_deb 'apt-cache policy do-agent | head -n3'
	exec_rpm 'yum --cacheonly list do-agent'
}

function command_create_status() {
	list \
		| jq -r '.droplets[] | "\(.id) [\(.name)] \(.status)"' \
		| GREP_COLOR='1;31' grep -P --color=yes 'new|$' \
		| GREP_COLOR='1;32' grep -P --color=yes 'active|$'
}

# scp a file to every host
function command_scp() {
	src=${1:-}; dest=${2:-}
	if [ -z "$src" ] || [ -z "$dest" ]; then
		abort "Usage: $0 scp <src> <dest>"
	fi

	for ip in $(list_ips); do
		# shellcheck disable=SC2029
		scp -o "UserKnownHostsFile=/dev/null" -o "StrictHostKeyChecking=no" -o "LogLevel=ERROR" "$src" root@"${ip}":"$dest" &
	done
	wait
}

# ssh to all debian-based droplets (ubuntu/debian) and execute a command
function exec_deb() {
	[ -z "$*" ] && abort "Usage: $0 exec_deb <command>"
	exec_ips "$(list_ips_deb)" "$*"
}

# ssh to all rpm-based droplets (centos/fedora) and execute a command
function exec_rpm() {
	[ -z "$*" ] && abort "Usage: $0 exec_rpm <command>"
	exec_ips "$(list_ips_rpm)" "$*"
}

# list all droplets without formatting
function list() {
	request GET "/droplets?tag_name=$TAG"
}

function list_ips() {
	list | jq -r '.droplets[].networks.v4[] | select(.type=="public") | .ip_address'
}

function list_ids() {
	list | jq -r '.droplets[].id'
}

function list_ips_deb() {
	list | \
		jq -r '.droplets[]
		| select(
			.image.distribution=="Debian"
			or
			.image.distribution=="Ubuntu"
		)
		| .networks.v4[]
		| select(.type=="public")
		| .ip_address'
}

function list_ips_rpm() {
	list | \
		jq -r '.droplets[]
		| select(
			.image.distribution=="CentOS"
			or
			.image.distribution=="Fedora"
		)
		| .networks.v4[]
		| select(.type=="public")
		| .ip_address'
}

# create a droplet with the provided image
function create_image() {
	image=$1
	if [ -z "$image" ]; then
		abort "Usage: ${FUNCNAME[0]} <image>"
	else
		echo "Creating image $image..."
	fi

	user_data=""

	if [ "${NO_INSTALL}" == "0" ]; then
		user_data=${USER_DATA_RPM}
		[[ "$image" =~ debian|ubuntu ]] && user_data=${USER_DATA_DEB}
	fi

	body=$(mktemp)
	cat <<EOF > "$body"
	{
		"name": "$image",
		"region": "nyc3",
		"size": "s-1vcpu-1gb",
		"image": "$image",
		"ssh_keys": [
			"${JONES_SSH_FINGERPRINT}",
			"${THOR_SSH_FINGERPRINT}",
			"${EVAN_SSH_FINGERPRINT}",
			"${SNYDER_SSH_FINGERPRINT}"
		],
		"backups": false,
		"ipv6": false,
		"user_data": "${user_data}",
		"tags": [ "${TAG}" ]
	}
EOF

	request POST "/droplets" "@${body}" \
		| jq -r '.droplet | "Created: \(.id): \(.name)"'

}


# Make an HTTP request to the API. The DATA param is optional
#
# Usage: request [METHOD] [PATH] [DATA]
#
# Examples:
#   request "GET" "/droplets"
#   request "POST" "/droplets" "@some-file.json"
#   request "POST" "/droplets" '{"some": "data"}'
#   request "DELETE" "/droplets/1234567"
function request() {
	[ -z "${AUTH_TOKEN:-}" ] && abort "AUTH_TOKEN is not set"

	METHOD=${1:-}
	URL=${2:-}
	DATA=${3:-}

	[ -z "$METHOD" ] && abort "Usage: ${FUNCNAME[0]} [METHOD] [PATH] [DATA]"

	if [[ ! "$URL" =~ ^/ ]] || [[ "$URL" =~ /v2 ]]; then
		abort "URL param should be a relative path not including v2 (e.g. /droplets). Got '$URL'"
	fi


	curl -SsL \
		-X "$METHOD" \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${AUTH_TOKEN}" \
		-d "$DATA" \
		"https://api.digitalocean.com/v2$URL"
}

# ask the user for input
function ask() {
	question=${1:-}
	[ -z "$question" ] && abort "Usage: ${FUNCNAME[0]} <question>"
	read -p "$question " -n 1 -r
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

# launch a uri with the system's default application (browser)
function launch() {
	uri=${1:-}
	[ -z "$uri" ] && abort "Usage: ${FUNCNAME[0]} <uri>"

	if [[ "$OS" =~ linux ]]; then
		xdg-open "$uri"
	else
		open "$uri"
	fi
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
}

# never put anything below this line. This is to prevent any partial execution
# if curl ever interrupts the download prematurely. In that case, this script
# will not execute since this is the last line in the script.
err_report() { echo "Error on line $1"; }
trap 'err_report $LINENO' ERR
main "$@"
