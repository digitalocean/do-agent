# DigitalOcean Agent

[![Build
Status](https://travis-ci.org/digitalocean/do-agent.svg?branch=master)](https://travis-ci.org/digitalocean/do-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/digitalocean/do-agent)](https://goreportcard.com/report/github.com/digitalocean/do-agent)
[![Coverage Status](https://coveralls.io/repos/github/digitalocean/do-agent/badge.svg?branch=feat%2Fadd-coveralls-report)](https://coveralls.io/github/digitalocean/do-agent?branch=feat%2Fadd-coveralls-report)

## Overview
do-agent enables droplet metrics to be gathered and sent to DigitalOcean to provide resource usage graphs and alerting. 

DO Agent currently supports:
- Ubuntu 14.04+
- Debian 8+
- Fedora 27+
- CentOS 6+
- Docker (see below)

## Installation

To install the do-agent on new Droplets simply select the Monitoring checkbox on the Droplet create screen to get the latest stable version of do-agent. Use your OS package manager (yum/dnf/apt-get) to update and manage do-agent.

## Installing via package managers

```bash
curl -sSL https://insights.nyc3.cdn.digitaloceanspaces.com/install.sh | sudo bash
# or wget
wget -qO- https://insights.nyc3.cdn.digitaloceanspaces.com/install.sh | sudo bash
```

If you prefer to inspect the script first:

```bash
curl -L -o ./install.sh https://insights.nyc3.cdn.digitaloceanspaces.com/install.sh
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
