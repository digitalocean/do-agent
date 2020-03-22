# DigitalOcean Agent

[![Build
Status](https://travis-ci.org/digitalocean/do-agent.svg?branch=master)](https://travis-ci.org/digitalocean/do-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/digitalocean/do-agent)](https://goreportcard.com/report/github.com/digitalocean/do-agent)
[![Coverage Status](https://coveralls.io/repos/github/digitalocean/do-agent/badge.svg?branch=master)](https://coveralls.io/github/digitalocean/do-agent?branch=master)

## Overview
do-agent enables droplet metrics to be gathered and sent to DigitalOcean to provide resource usage graphs and alerting. 

DO Agent currently supports:
- Ubuntu (oldest [End Of Standard Support](https://wiki.ubuntu.com/Releases) LTS release and later)
- Debian ([oldest supported](https://wiki.debian.org/LTS) LTS release and later)
- Fedora 27+
- CentOS 6+
- Docker (see below)

Note:

Although, we only officially support these distros and versions, do-agent works on most Linux distributions. Feel free to run it wherever you are successful, but any issues you encounter will not have official support from DigitalOcean

### Special Note For SELinux Users

The do-agent install script sets the `nis_enabled` flag to 1. Without this setting the do-agent cannot reach the network to perform authentication or send metrics to DigitalOcean backend servers. If you reverse this action, or install the do-agent on a machine manually you will need to run `setsebool -P nis_enabled 1 && systemctl daemon-reexec` otherwise the do-agent will not operate.

## Installation

To install the do-agent on new Droplets simply select the Monitoring checkbox on the Droplet create screen to get the latest stable version of do-agent. Use your OS package manager (yum/dnf/apt-get) to update and manage do-agent.

## Installing via package managers

```bash
curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo bash
# or wget
wget -qO- https://repos.insights.digitalocean.com/install.sh | sudo bash
```

If you prefer to inspect the script first:

```bash
curl -L -o ./install.sh https://repos.insights.digitalocean.com/install.sh
# inspect the file
less ./install.sh
# execute the file
sudo ./install.sh
```

## Development

### Requirements

- [go](https://golang.org/dl/) 1.11 or later
- [GNU Make](https://www.gnu.org/software/make/)
- Docker

```
git clone git@github.com:digitalocean/do-agent.git
cd do-agent

### build the project
make

### add dependencies
# first make sure you have the appropriate flags set to use go modules
# We recommend using https://github.com/direnv/direnv to automatically set
# these from the .envrc file in this project or you can manually set them
export GO111MODULE=on GOFLAGS=-mod=vendor

# then add your imports to any go file and run
go mod vendor
```

### Uninstall

do-agent can be uninstalled with your distribution's package manager

`apt-get remove do-agent` for Debian based distros

`yum remove do-agent` for RHEL based distros


### Run as a Docker container

You can optionally run do-agent as a docker container. In order to do so
you need to mount the host directory `/proc` to `/host/proc`.

For example:

```
docker run \
        -v /proc:/host/proc:ro \
        -v /sys:/host/sys:ro \
        digitalocean/do-agent:stable
```

## Report an Issue
Feel free to [open an issue](https://github.com/digitalocean/do-agent/issues/new)
if one does not [already exist](https://github.com/digitalocean/do-agent/issues)
