#!/bin/sh
# noexpandtab is required for EOF/heredoc
# vim: noexpandtab
#
# IMPORTANT: this script will execute with /bin/sh which is dash on some
# systems so this shebang should not be changed
# DO NOT change this and make sure you are linting with shellcheck to ensure
# compatbility with scripts

set -ue

SVC_NAME=do-agent
USERNAME=do-agent
INSTALL_DIR=/opt/digitalocean/${SVC_NAME}
CRON=/etc/cron.daily/${SVC_NAME}
INIT_SVC_FILE="/etc/init/${SVC_NAME}.conf"
SYSTEMD_SVC_FILE="/etc/systemd/system/${SVC_NAME}.service"

main() {
	update_selinux

	# fix for IN-2630
	# TODO remove userdel in next release
	# temporarily remove the do-agent and recreate it with no shell
	userdel -f do-agent || true
	useradd -s /bin/false -M --system $USERNAME || true

	# TODO remove this if in next release
	# temporarily download keys in a GPG keyring for debian based systems.
	if command -v apt-get >/dev/null 2>&1; then
		REPO_HOST=https://repos.insights.digitalocean.com
		REPO_GPG_KEY=${REPO_HOST}/sonar-agent.asc
		deb_list=/etc/apt/sources.list.d/digitalocean-agent.list
		deb_keyfile=/usr/share/keyrings/digitalocean-agent-keyring.gpg
		repo="do-agent"
		echo "deb [signed-by=${deb_keyfile}] ${REPO_HOST}/apt/${repo} main main" >"${deb_list}" || true
		curl -sL "${REPO_GPG_KEY}" | gpg --dearmor >"${deb_keyfile}" || true
	fi

	if command -v systemctl >/dev/null 2>&1; then
		# systemd is used, remove the upstart script
		rm -f "${INIT_SVC_FILE}"
		# systemctl enable --now is unsupported on older versions of debian/systemd
		echo "enable systemd service"
		systemctl daemon-reload
		systemctl enable -f ${SVC_NAME}
		systemctl restart ${SVC_NAME}
	elif command -v initctl >/dev/null 2>&1; then
		# upstart is used, remove the systemd script
		rm -f "${SYSTEMD_SVC_FILE}"
		echo "enable upstart service"
		initctl stop ${SVC_NAME} || true
		initctl reload-configuration
		initctl start ${SVC_NAME}
	else
		echo "Unknown init system. Exiting..." > /dev/stderr
		exit 1
	fi

	patch_updates
}


update_selinux() {
	echo "Detecting SELinux"
	enforced=$(getenforce 2>/dev/null || echo)

	if [ "$enforced" != "Enforcing" ]; then
		echo "SELinux not enforced"
		return
	fi

	echo "setting nis_enabled to 1 to allow ${SVC_NAME} to execute"
	setsebool -P nis_enabled 1 || echo "Failed" > /dev/stderr
	systemctl daemon-reexec || true
}

patch_updates() {
	# make sure we have the latest
	[ -f "${CRON}" ] && rm -f "${CRON}"
	script="${INSTALL_DIR}/scripts/update.sh"

	cat <<-EOF > "${CRON}"
	#!/bin/sh
	/bin/bash ${script} >/dev/null 2>&1
	EOF

	chmod +x "${CRON}"
}

main
