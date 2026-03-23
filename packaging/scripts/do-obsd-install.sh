#!/bin/bash
#
# This script is meant for quick & easy install via:
#   curl -sSL https://repos.insights.digitalocean.com/do-obsd-install.sh | sudo bash
# or:
#   wget -qO- https://repos.insights.digitalocean.com/do-obsd-install.sh | sudo bash
#
# vim: noexpandtab

set -ueo pipefail

REPO_HOST=https://repos.insights.digitalocean.com
REPO_GPG_KEY=${REPO_HOST}/sonar-agent.asc

repo="do-obsd-preview"
stable_repo="do-agent"

dist="unknown"
deb_list=/etc/apt/sources.list.d/digitalocean-obsd.list
deb_keyfile=/usr/share/keyrings/digitalocean-obsd-keyring.gpg
rpm_repo=/etc/yum.repos.d/digitalocean-obsd.repo

# do-agent stable channel sources (needed to resolve the do-agent dependency)
stable_deb_list=/etc/apt/sources.list.d/digitalocean-agent.list
stable_deb_keyfile=/usr/share/keyrings/digitalocean-agent-keyring.gpg
stable_rpm_repo=/etc/yum.repos.d/digitalocean-agent.repo

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
	wait_for_apt && ( apt-get purge -y do-obsd >/dev/null 2>&1 || : )

	echo "Installing apt repository..."
	wait_for_apt && ( apt-get -qq update || true )
	wait_for_apt && apt-get -qq install -y ca-certificates gnupg2 apt-utils apt-transport-https curl

	echo -n "Installing gpg key..."
	curl -sL "${REPO_GPG_KEY}" | gpg --dearmor >"${deb_keyfile}"

	# configure do-agent stable channel so the do-agent dependency resolves
	if [ ! -f "${stable_deb_list}" ]; then
		echo "Configuring do-agent stable channel..."
		cp "${deb_keyfile}" "${stable_deb_keyfile}"
		echo "deb [signed-by=${stable_deb_keyfile}] ${REPO_HOST}/apt/${stable_repo} main main" >"${stable_deb_list}"
		wait_for_apt && apt-get -qq update -o Dir::Etc::SourceParts=/dev/null -o APT::Get::List-Cleanup=no -o Dir::Etc::SourceList="sources.list.d/digitalocean-agent.list"
	fi

	# configure do-obsd preview channel
	echo "deb [signed-by=${deb_keyfile}] ${REPO_HOST}/apt/${repo} main main" >"${deb_list}"
	wait_for_apt && apt-get -qq update -o Dir::Etc::SourceParts=/dev/null -o APT::Get::List-Cleanup=no -o Dir::Etc::SourceList="sources.list.d/digitalocean-obsd.list"

	wait_for_apt && apt-get -qq install -y do-obsd
}

function install_rpm() {
	echo "Installing yum repository..."

	yum remove -y do-obsd || :

	yum install -y gpgme ca-certificates

	# configure do-agent stable channel so the do-agent dependency resolves
	if [ ! -f "${stable_rpm_repo}" ]; then
		echo "Configuring do-agent stable channel..."
		cat <<-STABLE_EOF > "${stable_rpm_repo}"
		[digitalocean-agent]
		name=DigitalOcean Agent
		baseurl=${REPO_HOST}/yum/${stable_repo}/\$basearch
		repo_gpgcheck=0
		gpgcheck=1
		enabled=1
		gpgkey=${REPO_GPG_KEY}
		sslverify=0
		sslcacert=/etc/pki/tls/certs/ca-bundle.crt
		metadata_expire=300
		STABLE_EOF
		yum --disablerepo="*" --enablerepo="digitalocean-agent" makecache
	fi

	# configure do-obsd preview channel
	cat <<-EOF > /etc/yum.repos.d/digitalocean-obsd.repo
	[digitalocean-obsd]
	name=DigitalOcean Observability Supervisor
	baseurl=${REPO_HOST}/yum/${repo}/\$basearch
	repo_gpgcheck=0
	gpgcheck=1
	enabled=1
	gpgkey=${REPO_GPG_KEY}
	sslverify=0
	sslcacert=/etc/pki/tls/certs/ca-bundle.crt
	metadata_expire=300
	EOF

	yum --disablerepo="*" --enablerepo="digitalocean-obsd" makecache
	yum install -y do-obsd
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
	read -r sys_vendor < /sys/devices/virtual/dmi/id/bios_vendor
	if ! [ "$sys_vendor" = "DigitalOcean" ]; then
		cat <<-EOF

		The DigitalOcean Observability Supervisor is only supported on DigitalOcean machines.

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
