#!/usr/bin/env bash

cat <<-HELPTEXT

 Upgrade/install the latest do-agent on all of the accounts droplets.
 You must provide a DO API token as AUTH_TOKEN environment variable.

 There is also a SSH_KEY environment variable that may be set to provide the location of an SSH identity file to be used.
 By default, ~/.ssh/id_rsa will be used if no SSH_KEY environment variable is provided

 Examples:
     AUTH_TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX ./upgrade-droplets.sh install
     AUTH_TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX SSH_KEY=~/.ssh/id_rsa ./upgrade-droplets.sh install

HELPTEXT

set -ue

SSH_KEY=${SSH_KEY:-"~/.ssh/id_rsa"}

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
				-i "${SSH_KEY}" \
				"${USER:-root}@${ip}" "${script}" 2>/dev/stdout || true
		)" &
	done
	wait
}

# install the latest stable version of the agent
function command_install() {
	exec_ips "$(list_ips)" "curl -SsL https://insights.nyc3.digitaloceanspaces.com/install.sh | sudo bash"
}

# list all droplets without formatting
function list() {
	request GET "/droplets"
}

function list_ips() {
	list | jq -r '.droplets[].networks.v4[] | select(.type=="public") | .ip_address'
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
