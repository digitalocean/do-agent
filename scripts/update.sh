#!/bin/bash
# vim: noexpandtab

main() {
	# add some jitter to prevent overloading the packaging machines
	sleep $(( RANDOM % 900 ))

	export DEBIAN_FRONTEND="noninteractive"
	if command -v apt-get 2&>/dev/null; then
		apt-get -qq update -o Dir::Etc::SourceParts=/dev/null -o APT::Get::List-Cleanup=no -o Dir::Etc::SourceList="sources.list.d/digitalocean-agent.list"
		apt-get -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" -qq install -y --only-upgrade do-agent
	elif command -v yum 2&>/dev/null; then
		yum -q -y --disablerepo="*" --enablerepo="digitalocean-agent" makecache
		yum -q -y update do-agent
	fi
}

main
