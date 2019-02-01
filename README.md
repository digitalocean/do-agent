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

### Deb Repository
```
echo "deb https://repos.sonar.digitalocean.com/apt main main" > /etc/apt/sources.list.d/digitalocean-agent.list
curl https://repos.sonar.digitalocean.com/sonar-agent.asc | sudo apt-key add -
sudo apt-get update
sudo apt-get install do-agent
```

### Yum Repository

```
cat <'EOF' > /etc/yum.repos.d/digitalocean-agent.repo
[sonar]
name=do agent
baseurl=https://repos.sonar.digitalocean.com/yum/$basearch
failovermethod=priority
enabled=1
gpgcheck=1
gpgkey=https://repos.sonar.digitalocean.com/sonar-agent.asc
EOF

rpm --import https://repos.sonar.digitalocean.com/sonar-agent.asc
yum install do-agent
```

## Development

### Requirements

- [go](https://golang.org/dl/)
- [golang/dep](https://github.com/golang/dep#installation)
- [GNU Make](https://www.gnu.org/software/make/)
- [gometalinter](https://github.com/alecthomas/gometalinter#installing)

```
git clone git@github.com:digitalocean/do-agent.git \
        $GOPATH/src/github.com/digitalocean/do-agent
cd !$

# build the project
make

# add dependencies
dep ensure -v -add <import path>
```

### Uninstall

do-agent can be uninstalled with your distribution's package manager

`apt remove do-agent` for Debian based distros

`yum remove do-agent` for RHEL based distros


### Run as a Docker container

You can optionally run do-agent as a docker container. In order to do so
you need to mount the host directory `/proc` to `/host/proc`.

For example:

```
docker run \
        -v /proc:/host/proc:ro \
        -v /sys:/host/sys:ro \
        digitalocean/do-agent:1
```

## Report an Issue
Feel free to [open an issue](https://github.com/digitalocean/do-agent/issues/new)
if one does not [already exist](https://github.com/digitalocean/do-agent/issues)
