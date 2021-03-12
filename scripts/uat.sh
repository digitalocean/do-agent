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
TAG=${TAG:-do-agent-uat-${USER}}

# map name=>imageID OR DO image name
# images should use IDs since images may be deleted from
# DO website
declare -A SUPPORTED_IMAGES
SUPPORTED_IMAGES["debian-9-x64"]="45446272"
SUPPORTED_IMAGES["debian-10-x64"]="69440038"
SUPPORTED_IMAGES["centos-7-x64"]="45446241"
SUPPORTED_IMAGES["centos-8-x64"]="69439535"
SUPPORTED_IMAGES["fedora-30-x64"]="46617047"
SUPPORTED_IMAGES["fedora-31-x64"]="69457626"
SUPPORTED_IMAGES["ubuntu-16-04-x32"]="45446271"
SUPPORTED_IMAGES["ubuntu-16-04-x64"]="45446373"
SUPPORTED_IMAGES["ubuntu-18-04-x64"]="45446242"
SUPPORTED_IMAGES["ubuntu-20-04-x64"]="62569011"

declare -a SSH_FINGERPRINTS=(
"4c:3d:0d:94:25:ba:00:b0:88:9f:d2:cc:97:8e:8f:43"
"47:31:9b:8b:87:a7:2d:26:79:17:87:83:53:65:d4:b4"
)

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
	grep -E '^function command_' "$0" \
		| sed 's,function command_,,g' \
		| sed 's,() {,,g' \
		| sort \
		| xargs -n1 -I{} echo "  {}"
	echo
}

# destroy all droplets tagged with $TAG
function command_destroy() {
	confirm "Are you sure you want to destroy all droplets with the tag ${TAG}?" \
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

function command_grafana() {
	urls=$(command_list_ids | xargs -n1 -I{} echo "https://grafana.internal.digitalocean.com/d/Sx4Cj7riz/insights-droplet-graphs?orgId=1&var-DropletID={}&var-region=nyc3&var-Datacenter=prod-pandora-droplets-nyc3&from=now-30m&to=now" | tee /dev/stderr)
	if confirm "Open these urls?"; then
		for u in $urls; do
			launch "$u" &>/dev/null
		done
	else
		echo "Aborting"
	fi
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
	if [ -n "$(list_ips)" ]; then
		abort "You already have a set of droplets created with this tag: ${TAG}. Either destroy them or try again with a different TAG"
	fi

	for name in "${!SUPPORTED_IMAGES[@]}"; do
		create_image "$name" &
	done
	wait
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

# check the cloud init status of a newly created droplet
function command_init_status() {
	command_exec "if [ -f /var/lib/cloud/instance/boot-finished ]; then \
		echo \"$(tput setaf 2)Ready $(tput sgr 0)\"
	else \
		echo \"$(tput setaf 1)NOT ready $(tput sgr 0)\"
	fi"
}

# ssh to all droplets and run yum/apt update to upgrade to the latest published
# version of do-agent
function command_update() {
	exec_ips "$(list_ips)" /opt/digitalocean/do-agent/scripts/update.sh
}

# remove do-agent from all uat machines
function command_uninstall() {
	exec_rpm "yum remove -qq -y do-agent"
	exec_deb "apt-get -qq purge -y do-agent"
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
				-o "BatchMode=yes" \
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
function command_version() {
	exec_deb 'apt-cache policy do-agent | head -n3'
	exec_rpm 'yum list -C do-agent'
}

function command_create_status() {
	list \
		| jq -r '.droplets[] | "\(.id) [\(.name)] \(.status)"' \
		| GREP_COLOR='1;31' grep -E --color=yes 'new|$' \
		| GREP_COLOR='1;32' grep -E --color=yes 'active|$'
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

# install a version of the agent. Can be unstable, beta, or stable
function command_install() {
	vers=${1:-}
	vers=${vers// /} # lowercase

	case "${vers// /}" in
		unstable)
			exec_ips "$(list_ips)" "curl -SsL https://repos.insights.digitalocean.com/install.sh | sudo UNSTABLE=1 bash"
			;;
		beta)
			exec_ips "$(list_ips)" "curl -SsL https://repos.insights.digitalocean.com/install.sh | sudo BETA=1 bash"
			;;
		stable)
			exec_ips "$(list_ips)" "curl -SsL https://repos.insights.digitalocean.com/install.sh | sudo bash"
			;;
		old)
			exec_ips "$(list_ips)" "curl -SsL https://agent.digitalocean.com/install.sh | sudo bash"
			;;
		*)
			abort "Usage: $0 install <unstable|beta|stable|old>"
			;;
	esac
}

function command_bin_version() {
	exec_ips "$(list_ips)" /opt/digitalocean/bin/do-agent --version
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
	image_name=$1
	if [ -z "${image_name}" ]; then
		abort "Usage: ${FUNCNAME[0]} <image_name>"
	else
		echo "Creating image ${image_name}..."
	fi

	image_id="${SUPPORTED_IMAGES[${image_name}]}"
	if [ -z "${image_id}" ]; then
		abort "${FUNCNAME[0]} unknown image name ${image_name}"
	fi


	monitoring="false"
	if [ "${NO_INSTALL}" == "0" ]; then
		monitoring="true"
	fi

	body=$(mktemp)
	cat <<-EOF > "$body"
	{
		"name": "${image_name}",
		"monitoring": "${monitoring}",
		"region": "nyc3",
		"size": "s-1vcpu-1gb",
		"image": "${image_id}",
		"ssh_keys": $(printf "\"%s\"" "${SSH_FINGERPRINTS[@]}" | jq -s '.'),
		"backups": false,
		"ipv6": false,
		"tags": [ "${TAG}", "agent-testing" ]
	}
	EOF

	echo "Image: ${image_name}: $( request POST "/droplets" "@${body}" \ | jq -r '.droplet | "ID: \(.id)"')"
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
