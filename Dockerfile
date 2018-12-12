FROM ubuntu:18.04
MAINTAINER Insights Engineering <eng-insights@digitalocean.com>

RUN set -x && \
        apt-get -qq update && \
        apt-get install -y ca-certificates && \
        apt-get autoclean

ADD target/do-agent-linux-amd64 /bin/do-agent

RUN mkdir -p /host

VOLUME /host/proc
VOLUME /host/sys

ENTRYPOINT ["/bin/do-agent", "--path.procfs", "/host/proc", "--path.sysfs", "/host/sys"]
