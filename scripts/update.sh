#!/bin/bash
# vim: noexpandtab

main() {
	if command -v apt-get 2&>/dev/null; then
		apt-get -qq update -o Dir::Etc::sourcelist="sources.list.d/digitalocean-agent.list" -o Dir::Etc::sourceparts="-" -o APT::Get::List-Cleanup="0"
		apt-get -qq install -y --only-upgrade do-agent
	elif command -v yum 2&>/dev/null; then
		yum -q -y update do-agent
	fi
}

main
