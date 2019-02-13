FROM ubuntu:18.04

RUN set -x && \
        apt-get -qq update && \
        apt-get install -y ca-certificates dumb-init && \
        apt-get autoclean

ADD target/do-agent-linux-amd64 /bin/do-agent

RUN mkdir -p /host

VOLUME /host/proc
VOLUME /host/sys

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/bin/do-agent", "--path.procfs", "/host/proc", "--path.sysfs", "/host/sys"]
