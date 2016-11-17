# Do-Agent

[![Build Status](https://travis-ci.org/digitalocean/do-agent.svg?branch=master)](https://travis-ci.org/digitalocean/do-agent)
[![GoDoc](https://godoc.org/github.com/digitalocean/do-agent?status.svg)](https://godoc.org/github.com/digitalocean/do-agent)

## General

The `do-agent` extracts system metrics from
DigitalOcean Droplets and transmits them to the DigitalOcean
monitoring service. When the agent is initiated on a Droplet it
will automatically configure itself with the appropriate settings.

## Flags

Flag |Option |Description
-----|-------|-----------
-log_syslog | bool | Log to syslog.
-log_level | string | Sets the log level. [ INFO, ERROR, DEBUG ] Default=INFO
-force_update | bool | force an agent update.
-v | bool | Prints DigitalOcean Agent version.
-h | bool | Prints `do-agent` usage information.

## Env Flags

Generally used during development and debugging.

Flag |Option |Description
-----|-------|-----------
DO_AGENT_AUTHENTICATION_URL | string | Override the authentication URL
DO_AGENT_APPKEY | string | Override AppKey
DO_AGENT_METRICS_URL | string | Override metrics URL
DO_AGENT_DROPLET_ID | int64 | Override Droplet ID
DO_AGENT_AUTHTOKEN | string | Override AuthToken
DO_AGENT_UPDATE_URL | string | Override Update URL
DO_AGENT_REPO_PATH | string | Override Local repository path
DO_AGENT_PLUGIN_PATH | string | Override plugin directory path

## Building and running

    `make build`
    `sudo -u nobody ./do-agent <flags>`

## Running Tests

    `make test`

## Installing via package managers

### Deb Repository
```
echo "deb https://repos.sonar.digitalocean.com/apt main main" > /etc/apt/sources.list.d/sonar.list
curl https://repos.sonar.digitalocean.com/sonar-agent.asc | sudo apt-key add -
sudo apt-get update
sudo apt-get install do-agent
```

### Yum Repository

```
cat <'EOF' > /etc/yum.repos.d/DigitalOcean-Sonar.repo
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

## Update via package managers

### Yum

`yum update do-agent`


### Apt

`apt-get update && apt-get install --only-upgrade do-agent`


## Removal via package managers

### Yum

`sudo yum remove do-agent`


### Apt

`sudo apt-get purge do-agent`


## Package installation

### Deb

`dpkg -i do-agent_<version>_<build>.deb`


### Rpm

`rpm -Uvh do-agent-<version>_<build>.rpm`


## Package removal

### Deb

`dpkg --remove do-agent`


### Rpm

`rpm -e do-agent`


## Plugins

`do-agent` builds in a common set of metrics, and provides a plugin mechanism to
add additional metric collectors.  Plugins are executable files that are placed
in the agent's plugin directory (see `DO_AGENT_PLUGIN_PATH`).  When `do-agent`
starts, it will find all executables in the plugin path and call them during
metric collection.

A collection agent must do two things:

1. report metric configuration to stdout when a "config" argument is passed.
1. report metric values to stdout when no argument is passed.

See `plugins/test.sh` for a simple static plugin, in the form of a shell script.
Plugins may be written in any language, and must only know how to produce
serialized results in `json` format.


### Definitions

When test.sh is run with "config" as the argument, it produces the following:

```json
{
  "definitions": {
    "test": {
      "type": 1,
      "labels": {
        "user": "foo"
      },
      "label_keys": ["timezone", "country"]
    }
  }
}
```

Definitions is a mapping of metric name to configuration.  `type` is an integer
type, compatible with the Prometheus client definition.  The following types are
supported:

Type Value |Description
---|--------
0  | Counter
1  | Gauge

Two types of labels are supported: fixed and dynamic.  Fixed labels are specified
once in the metric definition and never specified in the metric value.  In the
example above, there is a fixed label `user:foo`.  This label will be set on
all values of the metric.

The second type of label is a dynamic label.  Dynamic labels have a fixed key
but the value is specified with every update of the metric.  The example above
has two dynamic labels, with the keys `timezone` and `country`.  The values
for these labels will be provided on each metric update.

### Values

When test.sh is run without any arguments, it produces the current value of
the collected metrics.  For example:

```json
{
  "metrics": {
    "test": {
      "value": 42.0,
      "label_values": ["est", "usa"]
    }
  }
}
```

The _(optional)_ label_values must correspond to the label_keys in the definition.
If there were two label keys in the definition, then there must be exactly two
label values in each metric update.

## Contributing

The `do-agent` project makes use of the [GitHub Flow](https://guides.github.com/introduction/flow/)
for contributions.

If you'd like to contribute to the project, please
[open an issue](https://github.com/digitalocean/do-agent/issues/new) or find an
[existing issue](https://github.com/digitalocean/do-agent/issues) that you'd like
to take on.  This ensures that efforts are not duplicated, and that a new feature
aligns with the focus of the rest of the repository.

Once your suggestion has been submitted and discussed, please be sure that your
code meets the following criteria:
  - code is completely `gofmt`'d
  - new features or codepaths have appropriate test coverage
  - `go test ./...` passes
  - `go vet ./...` passes
  - `golint ./...` returns no warnings, including documentation comment warnings

In addition, if this is your first time contributing to the `do-agent` project,
add your name and email address to the
[AUTHORS](https://github.com/digitalocean/do-agent/blob/master/AUTHORS) file
under the "Contributors" section using the format:
`First Last <email@example.com>`.

Finally, submit a pull request for review!

## Report a bug

If you discover a software bug, please feel free to report it by:

1. Search through the [existing issues](https://github.com/digitalocean/do-agent/issues) to ensure that the bug hasn't already been filed.
2. [Open an issue](https://github.com/digitalocean/do-agent/issues/new) for a new bug.
