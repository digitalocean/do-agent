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

main() {
	create_user

	if command -v systemctl >/dev/null 2>&1; then
		echo "Stopping systemctl service..."
		systemctl stop ${SVC_NAME} 2>/dev/null || true
	elif command -v initctl >/dev/null 2>&1; then
		echo "Stopping initctl service..."
		initctl stop ${SVC_NAME} 2>/dev/null || true
	else
		echo "ERROR: Unknown init system" > /dev/stderr
	fi
}

create_user() {
	# create the user if it doesn't already exist
	if ! getent passwd $USERNAME >/dev/null 2>&1; then
		echo "Creating $USERNAME user..."
		adduser --system $USERNAME
	fi
}

main
