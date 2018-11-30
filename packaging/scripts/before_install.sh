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

main () {
	if command -v systemctl >/dev/null 2>&1; then
		systemctl stop ${SVC_NAME} || true
	elif command -v initctl >/dev/null 2>&1; then
		initctl stop ${SVC_NAME} || true
	else
		echo "Unknown init system" > /dev/stderr
	fi
}

main
