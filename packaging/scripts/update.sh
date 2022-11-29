#!/bin/bash
# vim: noexpandtab

REPO_HOST=https://repos.insights.digitalocean.com
REPO_GPG_KEY_CURRENT=${REPO_HOST}/sonar-agent-current.asc

main() {
	# add some jitter to prevent overloading the packaging machines
	sleep $(( RANDOM % 900 ))

	export DEBIAN_FRONTEND="noninteractive"
	if command -v apt-get 2&>/dev/null; then
		curl -sL "${REPO_GPG_KEY_CURRENT}" | apt-key add -
		apt-get -qq update -o Dir::Etc::SourceParts=/dev/null -o APT::Get::List-Cleanup=no -o Dir::Etc::SourceList="sources.list.d/digitalocean-agent.list"
		apt-get -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" -qq install -y --only-upgrade do-agent
	elif command -v yum 2&>/dev/null; then
		rpm --import "${REPO_GPG_KEY_CURRENT}"
		sed -i 's/gpgkey=https:\/\/repos.insights.digitalocean.com\/sonar-agent.asc/gpgkey=https:\/\/repos.insights.digitalocean.com\/sonar-agent-current.asc/' /etc/yum.repos.d/digitalocean-agent.repo
		yum -q -y --disablerepo="*" --enablerepo="digitalocean-agent" makecache
		yum -q -y update do-agent
	fi
}

main
