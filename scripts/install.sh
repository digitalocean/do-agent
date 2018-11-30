#!/usr/bin/env bash

set -ueo pipefail

# REPO="do-agent"
# TODO use metadata api to determine beta
REPO="do-agent-beta"
dist=""

function main() {
	check_dist
	kind=$( { [[ "$dist" =~ debian|ubuntu ]] && echo "deb"; } \
		|| echo "rpm")

	require_package curl

	curl -s "https://packagecloud.io/install/repositories/digitalocean-insights/${REPO}/script.$kind.sh" \
		| sudo bash

	install_package do-agent
}

function install_package() {
	[ -z "${1:-}" ] && \
		abort "Usage: ${FUNCNAME[0]} <package>"

	case "$dist" in
		debian|ubuntu)
			apt-get install -q -y "$1"
			;;
		centos|fedora)
			yum -q -y install "$1"
			;;
		*)
			not_supported
			;;
	esac
}

function update_package_info() {
	case "$dist" in
		debian|ubuntu)
			apt-get update
			;;
		centos|fedora)
			return
			;;
		*)
			not_supported
			;;
	esac
}


function check_dist() {
	if [  -f /etc/os-release  ]; then
		dist=$(awk -F= '$1 == "ID" {gsub("\"", ""); print$2}' /etc/os-release)
	elif [ -f /etc/redhat-release ]; then
		dist=$(awk '{print tolower($1)}' /etc/redhat-release)
	else
		not_supported
	fi
}

function not_supported() {
	abort "unsupported distribution. If you feel this is an error contact support@digitalocean.com"
}

function require_package() {
	[ -z "${1:-}" ] && abort "Usage: ${FUNCNAME[0]} <package>"

	pkg="$1"

	if ! command -v "$pkg" 2&> /dev/null; then
		update_package_info
		install_package "$pkg"
	fi
}

function abort() {
	read -r line func file <<< "$(caller 0)"
	echo "ERROR in $file.$func:$line: $1" > /dev/stderr
	exit 1
}

# leave this last to prevent any partial executions
main
