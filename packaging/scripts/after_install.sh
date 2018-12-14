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
NOBODY_USER=nobody
NOBODY_GROUP=nogroup
CRON=/etc/cron.daily/do-agent

# fedora uses nobody instead of nogroup
getent group nobody 2> /dev/null \
	&& NOBODY_GROUP=nobody

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
}

patch_updates() {
	# make sure we have the latest
	[ -f "${CRON}" ] && rm -rf "${CRON}"
	script="${INSTALL_DIR}/scripts/update.sh"

	cat <<-EOF > "${CRON}"
	#!/bin/sh
	/bin/bash ${script}
	EOF

	chmod a+x "${CRON}"

	echo "cron installed"
}

init_systemd() {
	# cannot use symlink because of an old bug https://bugzilla.redhat.com/show_bug.cgi?id=955379
	SVC=/etc/systemd/system/${SVC_NAME}.service
	cat <<-EOF > "$SVC"
	[Unit]
	Description=DigitalOcean do-agent agent
	After=network-online.target
	Wants=network-online.target

	[Service]
	User=${NOBODY_USER}
	Group=${NOBODY_GROUP}
	ExecStart=/usr/local/bin/do-agent
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

	# enable --now is unsupported on older versions of debian/systemd
	systemctl enable ${SVC}
	systemctl stop ${SVC_NAME} || true
	systemctl start ${SVC_NAME}
}

init_upstart() {
	cat <<-EOF > /etc/init/${SVC_NAME}.conf
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
	  exec su -s /bin/sh -c 'exec "\$0" "\$@"' ${NOBODY_USER} -- /usr/local/bin/do-agent --syslog
	end script
	EOF
	initctl reload-configuration
	initctl stop ${SVC_NAME} || true
	initctl start ${SVC_NAME}
}


dist() {
	if [  -f /etc/os-release  ]; then
		awk -F= '$1 == "ID" {gsub("\"", ""); print$2}' /etc/os-release
	elif [ -f /etc/redhat-release ]; then
		awk '{print tolower($1)}' /etc/redhat-release
	fi
}


# never put anything below this line. This is to prevent any partial execution
# if curl ever interrupts the download prematurely. In that case, this script
# will not execute since this is the last line in the script.
main
