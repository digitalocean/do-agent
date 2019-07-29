#!/bin/sh
# noexpandtab is required for EOF/heredoc
# vim: noexpandtab
#
# IMPORTANT: this script will execute with /bin/sh which is dash on some
# systems so this shebang should not be changed
# DO NOT change this and make sure you are linting with shellcheck to ensure
# compatbility with scripts

set -ue

INSTALL_DIR=/opt/digitalocean/do-agent
SVC_NAME=do-agent
USERNAME=do-agent
CRON=/etc/cron.daily/do-agent
INIT_SVC_FILE="/etc/init/${SVC_NAME}.conf"
SYSTEMD_SVC_FILE="/etc/systemd/system/${SVC_NAME}.service"

main() {
	update_selinux

	useradd -M --system $USERNAME || true

	if command -v systemctl >/dev/null 2>&1; then
		# systemd is used, remove the upstart script
		rm -f "${INIT_SVC_FILE}"
	elif command -v initctl >/dev/null 2>&1; then
		# upstart is used, remove the systemd script
		rm -f "${SYSTEMD_SVC_FILE}"
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

	echo "setting nis_enabled to 1 to allow do-agent to execute"
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
