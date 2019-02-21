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
USERGROUP=nogroup
CRON=/etc/cron.daily/do-agent
INIT_SVC_FILE="/etc/init/${SVC_NAME}.conf"
SYSTEMD_SVC_FILE="/etc/systemd/system/${SVC_NAME}.service"

# fedora uses nobody instead of nogroup
getent group nobody 2> /dev/null \
	&& USERGROUP=nobody

main() {
	update_selinux

	if command -v systemctl >/dev/null 2>&1; then
		init_systemd
	elif command -v initctl >/dev/null 2>&1; then
		init_upstart
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
	[ -f "${CRON}" ] && rm -fv "${CRON}"
	script="${INSTALL_DIR}/scripts/update.sh"

	cat <<-EOF > "${CRON}"
	#!/bin/sh
	/bin/bash ${script}
	EOF

	chmod +x "${CRON}"
}

init_systemd() {
	echo "Creating ${SYSTEMD_SVC_FILE}..."
	# cannot use symlink because of an old bug https://bugzilla.redhat.com/show_bug.cgi?id=955379
	cat <<-EOF > "${SYSTEMD_SVC_FILE}"
	[Unit]
	Description=DigitalOcean do-agent agent
	After=network-online.target
	Wants=network-online.target

	[Service]
	User=${USERNAME}
	Group=${USERGROUP}
	ExecStart=/opt/digitalocean/bin/do-agent
	Restart=always

	OOMScoreAdjust=-900
	SyslogIdentifier=DigitalOceanAgent
	PrivateTmp=yes
	ProtectSystem=full
	ProtectHome=yes
	NoNewPrivileges=yes

	[Install]
	WantedBy=multi-user.target
	EOF

	# --now is unsupported on older versions of debian/systemd
	systemctl daemon-reload
	systemctl enable -f ${SVC_NAME}
	systemctl start ${SVC_NAME}
}

init_upstart() {
	echo "Creating ${INIT_SVC_FILE}..."
	cat <<-EOF > ${INIT_SVC_FILE}
	# do-agent - An agent that collects system metrics.
	#
	# An agent that collects system metrics and transmits them to DigitalOcean.
	description "The DigitalOcean Monitoring Agent"
	author "DigitalOcean"

	start on runlevel [2345]
	stop on runlevel [!2345]
	console none
	normal exit 0 TERM
	kill timeout 5
	respawn

	script
	  exec su -s /bin/sh -c 'exec "\$0" "\$@"' ${USERNAME} -- /opt/digitalocean/bin/do-agent --syslog
	end script
	EOF

	initctl reload-configuration
	initctl start ${SVC_NAME}
}

# never put anything below this line. This is to prevent any partial execution
# if curl ever interrupts the download prematurely. In that case, this script
# will not execute since this is the last line in the script.
main
