#!/bin/bash
#
# This script is meant for quick & easy install via:
#   curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo bash
# or:
#   wget -qO- https://repos.insights.digitalocean.com/install.sh | sudo bash
#
# To use the BETA branch of do-agent pass the BETA=1 flag to the script
#   curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo BETA=1 bash
#

set -ueo pipefail

UNSTABLE=${UNSTABLE:-0}
BETA=${BETA:-0}

REPO_HOST=https://repos.insights.digitalocean.com
REPO_GPG_KEY=${REPO_HOST}/sonar-agent.asc

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

	case "${dist}" in
		debian|ubuntu)
			install_apt
			;;
		centos|cloudlinux|fedora|almalinux|rocky)
			install_rpm
			;;
		*)
			not_supported
			;;
	esac
}

function wait_for_apt() {
	while fuser /var/{lib/{dpkg,apt/lists},cache/apt/archives}/lock >/dev/null 2>&1; do
		echo "Waiting on apt.."
		sleep 2
	done
}

function install_apt() {
	export DEBIAN_FRONTEND=noninteractive
	# forcefully remove any existing installations
	wait_for_apt && ( apt-get purge -y do-agent >/dev/null 2>&1 || : )

	echo "Installing apt repository..."
	wait_for_apt && ( apt-get -qq update || true )
	wait_for_apt && apt-get -qq install -y ca-certificates gnupg2 apt-utils apt-transport-https curl
	echo "deb ${REPO_HOST}/apt/${repo} main main" > /etc/apt/sources.list.d/digitalocean-agent.list
	echo -n "Installing gpg key..."
	curl -sL "${REPO_GPG_KEY}" | apt-key add -
	wait_for_apt && apt-get -qq update -o Dir::Etc::SourceParts=/dev/null -o APT::Get::List-Cleanup=no -o Dir::Etc::SourceList="sources.list.d/digitalocean-agent.list"
	wait_for_apt && apt-get -qq install -y do-agent
}

function install_rpm() {
	echo "Installing yum repository..."

	# forcefully remove any existing installations
	yum remove -y do-agent || :

	yum install -y gpgme ca-certificates

	cat <<-EOF > /etc/yum.repos.d/digitalocean-agent.repo
	[digitalocean-agent]
	name=DigitalOcean Agent
	baseurl=${REPO_HOST}/yum/${repo}/\$basearch
	repo_gpgcheck=0
	gpgcheck=1
	enabled=1
	gpgkey=${REPO_GPG_KEY}
	sslverify=0
	sslcacert=/etc/pki/tls/certs/ca-bundle.crt
	metadata_expire=300
	EOF

	yum --disablerepo="*" --enablerepo="digitalocean-agent" makecache
	yum install -y do-agent
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

function check_dist() {
	echo -n "Verifying compatibility with script..."
	if [  -f /etc/os-release  ]; then
		dist=$(awk -F= '$1 == "ID" {gsub("\"", ""); print$2}' /etc/os-release)
	elif [ -f /etc/redhat-release ]; then
		dist=$(awk '{print tolower($1)}' /etc/redhat-release)
	else
		not_supported
	fi

	dist=$(echo "${dist}" | tr '[:upper:]' '[:lower:]')

	case "${dist}" in
		debian|ubuntu|centos|fedora|rocky)
			echo "OK"
			;;
		cloudlinux|almalinux)
			echo "WARN ${dist} is not officially supported. Attempting RPM install"
			;;
		*)
			not_supported
			;;
	esac
}


function check_do() {
	echo -n "Verifying machine compatibility..."
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


# leave this last to prevent any partial execution
main
