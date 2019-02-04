#!/bin/bash

set -ueo pipefail

UNSTABLE=${UNSTABLE:-0}
BETA=${BETA:-0}

repo="do-agent"
[ "${UNSTABLE}" != 0 ] && repo="do-agent-unstable"
[ "${BETA}" != 0 ] && repo="do-agent-beta"

dist="unknown"
deb_list=/etc/apt/sources.list.d/digitalocean-agent.list
rpm_repo=/etc/yum.repos.d/digitalocean-agent.repo

function main() {
	[ "$(id -u)" != "0" ] && \
		abort "This script must be executed as root."

	clean
	check_do
	check_dist
	update_packages
	install_package curl

	kind=""
	case "${dist}" in
		debian|ubuntu)
			kind="deb"
			;;
		centos|fedora)
			kind="rpm"
			;;
		*)
			not_supported
			;;
	esac

	curl -sL "https://packagecloud.io/install/repositories/digitalocean-insights/${repo}/script.${kind}.sh" \
		| sudo bash

	move_source_file

	update_packages
	remove_package do-agent
	install_package do-agent
}

function clean() {
	echo -n "Cleaning up old sources..."
	if [ -f "$deb_list" ]; then
		rm -f "${deb_list}"
	elif [ -f "$rpm_repo" ]; then
		rm -f "${rpm_repo}"
	fi
	echo "OK"
}

function update_packages() {
	echo -n "Updating package caches..."
	case "${dist}" in
		debian|ubuntu)
			apt-get -qq clean || :
			apt-get -qq autoclean || :
			apt-get -qq update || :
			;;
		centos|fedora)
			yum -q clean all || :
			yum -q -y makecache || :
			;;
		*)
			not_supported
			;;
	esac
	echo "OK"
}

function remove_package() {
	pkg=${1:-}
	[ -z "${pkg}" ] && abort "Usage: ${FUNCNAME[0]} <package>"
	echo "Removing package: ${pkg}..."

	case "${dist}" in
		debian|ubuntu)
			apt-get -qq remove -y "${pkg}" || :
			;;
		centos|fedora)
			yum -q -y remove "${pkg}" || :
			;;
		*)
			not_supported
			;;
	esac
}

function install_package() {
	pkg=${1:-}
	[ -z "${pkg}" ] && abort "Usage: ${FUNCNAME[0]} <package>"
	echo "Installing package: ${pkg}..."

	case "${dist}" in
		debian|ubuntu)
			apt-get -qq install -y "${pkg}"
			;;
		centos|fedora)
			yum -q -y install "${pkg}"
			;;
		*)
			not_supported
			;;
	esac
}

# we want the file to be consistently named across the unstable, beta, and stable releases
function move_source_file() {
	echo "Renaming source file(s)..."
	src=""
	dest=""
	case "${dist}" in
		debian|ubuntu)
			src="/etc/apt/sources.list.d/digitalocean-insights_${repo}.list"
			dest=${deb_list}
			;;
		centos|fedora)
			src="/etc/yum.repos.d/digitalocean-insights_${repo}.repo"
			dest=${rpm_repo}
			;;
		*)
			not_supported
			;;
	esac

	mv -v "${src}" "${dest}"
	# use a consistent name for all versions of the repository
	sed -i "s,digitalocean-insights_${repo},digitalocean-agent,g" "${dest}"
}

function check_dist() {
	echo -n "Verifying compatability with script..."
	if [  -f /etc/os-release  ]; then
		dist=$(awk -F= '$1 == "ID" {gsub("\"", ""); print$2}' /etc/os-release)
	elif [ -f /etc/redhat-release ]; then
		dist=$(awk '{print tolower($1)}' /etc/redhat-release)
	else
		not_supported
	fi

	dist=$(echo "${dist}" | tr '[:upper:]' '[:lower:]')

	case "${dist}" in
		debian|ubuntu|centos|fedora)
			echo "OK"
			;;
		*)
			not_supported
			;;
	esac
}


function check_do() {
	echo -n "Verifying machine compatability..."
	# DigitalOcean embedded platform information in the DMI data.
	read -r sys_vendor < /sys/devices/virtual/dmi/id/bios_vendor
	if ! [ "$sys_vendor" = "DigitalOcean" ]; then
		cat <<-EOF

    The DigitalOcean Agent is only supported on DigitalOcean machines.

    If you are seeing this message on an older droplet, you may need to power-off
    and then power-on at http://cloud.digitalocean.com. After power-cycling,
    please re-run this script.

		EOF
		exit 1
	fi
	echo "OK"
}

function not_supported() {
	cat <<-EOF 

	This script does not support the OS/Distribution on this machine.
	If you feel that this is an error contact support@digitalocean.com
	or create an issue at https://github.com/digitalocean/do-agent/issues/new.

	EOF
	exit 1
}

# abort with an error message
function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file:$func:$line: $1" > /dev/stderr
	exit 1
}


# leave this last to prevent any partial executions
main
