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
CRON=/etc/cron.daily/do-agent

main() {
	if command -v systemctl >/dev/null 2>&1; then
		echo "Configure systemd..."
		clean_systemd
	elif command -v initctl >/dev/null 2>&1; then
		echo "Configure upstart..."
		clean_upstart
	else
		echo "Unknown init system" > /dev/stderr
	fi

	remove_cron
}

remove_cron() {
	rm -f "${CRON}"
	echo "cron removed"
}

clean_upstart() {
	initctl stop ${SVC_NAME} || true
	unlink /etc/init/${SVC_NAME}.conf || true
	initctl reload-configuration || true
}

clean_systemd() {
	systemctl stop ${SVC_NAME} || true
	systemctl disable ${SVC_NAME}.service || true
	unlink /etc/systemd/system/${SVC_NAME}.service || true
	systemctl daemon-reload || true
}

main
