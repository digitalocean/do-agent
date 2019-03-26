#!/bin/bash
# vim: noexpandtab

main() {
	if command -v apt-get 2&>/dev/null; then
		apt-get -qq update -o Dir::Etc::sourcelist="sources.list.d/digitalocean-agent.list"
		apt-get -qq install -y --only-upgrade do-agent
	elif command -v yum 2&>/dev/null; then
		yum --disablerepo="*" --enablerepo="digitalocean-agent" makecache
		yum -q -y update do-agent
	fi
}

main
