#!/bin/sh
# noexpandtab is required for EOF/heredoc
# vim: noexpandtab
#
# IMPORTANT: this script will execute with /bin/sh which is dash on some
# systems so this shebang should not be changed

set -ue

SVC_NAME=do-obsd
CRON=/etc/cron.daily/do-obsd
POLKIT_RULE=/etc/polkit-1/rules.d/60-do-obsd.rules

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
	else
		echo "Unknown init system" > /dev/stderr
	fi

	remove_polkit_rule
	remove_cron
}

clean_systemd() {
	echo "Cleaning up systemd scripts"
	systemctl stop ${SVC_NAME} || true
	systemctl disable ${SVC_NAME}.service || true
	systemctl daemon-reload || true
}

remove_polkit_rule() {
	echo "Removing polkit rule"
	rm -f "${POLKIT_RULE}"
}

remove_cron() {
	rm -fv "${CRON}"
}

main
