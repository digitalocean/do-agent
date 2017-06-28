FROM debian:9

RUN apt-get -qqy update && \
    apt-get -qqy install curl

RUN mkdir -p /agent
RUN mkdir -p /agent/updates
RUN mkdir -p /agent/proc

COPY build/do-agent_linux_amd64 /agent

ENV DO_AGENT_REPO_PATH   /agent/updates
ENV DO_AGENT_PROCFS_ROOT /agent/proc

CMD /agent/do-agent_linux_amd64
