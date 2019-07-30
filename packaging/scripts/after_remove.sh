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

# fix an issue where this script runs on upgrades for rpm
# see https://github.com/jordansissel/fpm/issues/1175#issuecomment-240086016
arg="${1:-0}"

main() {
	if echo "${arg}" | grep -qP '^\d+$' && [ "${arg}" -gt 0 ]; then
		# rpm upgrade
		exit 0
	elif echo "${arg}" | grep -qP '^upgrade$'; then
		# deb upgrade
		exit 0
	fi

	if command -v systemctl >/dev/null 2>&1; then
		clean_systemd
	elif command -v initctl >/dev/null 2>&1; then
		clean_upstart
	else
		echo "Unknown init system" > /dev/stderr
	fi

	remove_cron
}

remove_cron() {
	rm -fv "${CRON}"
}

clean_upstart() {
	echo "Cleaning up init scripts"
	initctl stop ${SVC_NAME} || true
	rm -fv /etc/init/${SVC_NAME}.conf || true
	initctl reload-configuration || true
}

clean_systemd() {
	echo "Cleaning up systemd scripts"
	systemctl stop ${SVC_NAME} || true
	systemctl disable ${SVC_NAME}.service || true
	rm -fv /etc/systemd/system/${SVC_NAME}.service || true
	systemctl daemon-reload || true
}

main
