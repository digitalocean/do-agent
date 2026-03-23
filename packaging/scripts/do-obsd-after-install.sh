#!/bin/sh
# noexpandtab is required for EOF/heredoc
# vim: noexpandtab
#
# IMPORTANT: this script will execute with /bin/sh which is dash on some
# systems so this shebang should not be changed

set -ue

SVC_NAME=do-obsd
INSTALL_DIR=/opt/digitalocean/${SVC_NAME}
CRON=/etc/cron.daily/${SVC_NAME}
POLKIT_RULE=/etc/polkit-1/rules.d/60-do-obsd.rules

main() {
	update_selinux
	install_polkit_rule
	enable_service
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

install_polkit_rule() {
	echo "Installing polkit rule for do-otelcol.service management"
	mkdir -p /etc/polkit-1/rules.d
	cat > "${POLKIT_RULE}" <<'POLKIT_EOF'
polkit.addRule(function(action, subject) {
    if (action.id == "org.freedesktop.systemd1.manage-units" &&
        action.lookup("unit") == "do-otelcol.service" &&
        subject.user == "do-agent") {
        return polkit.Result.YES;
    }
});
POLKIT_EOF
}

enable_service() {
	if command -v systemctl >/dev/null 2>&1; then
		echo "enable systemd service"
		systemctl daemon-reload
		systemctl enable -f ${SVC_NAME}
		systemctl restart ${SVC_NAME}
	else
		echo "Unknown init system. Exiting..." > /dev/stderr
		exit 1
	fi
}

patch_updates() {
	[ -f "${CRON}" ] && rm -f "${CRON}"
	script="${INSTALL_DIR}/scripts/update.sh"

	cat <<-EOF > "${CRON}"
	#!/bin/sh
	/bin/bash ${script} >/dev/null 2>&1
	EOF

	chmod +x "${CRON}"
}

main
