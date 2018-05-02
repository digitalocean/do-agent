FROM golang:1.9-alpine

ENV CGO=0
ENV GOOS=linux

ARG CURRENT_BRANCH
ARG CURRENT_HASH
ARG LAST_RELEASE

RUN  apk update && \
     apk add bash && \
     apk add curl && \
     apk add git && \
     apk add make && \
     apk add libc6-compat

COPY . /go/src/github.com/digitalocean/do-agent

RUN cd /go/src/github.com/digitalocean/do-agent && \
    set -x && \
    make build RELEASE=${LAST_RELEASE} CURRENT_BRANCH=${CURRENT_BRANCH} CURRENT_HASH=${CURRENT_HASH}

# Copy what is needed to
FROM alpine
ENV DO_AGENT_REPO_PATH   /agent/updates
ENV DO_AGENT_PROCFS_ROOT /agent/proc

RUN mkdir -p /agent
RUN mkdir -p /agent/updates
RUN mkdir -p /agent/proc

RUN  apk update && \
     apk add libc6-compat && \
     apk add ca-certificates

COPY --from=0 /go/src/github.com/digitalocean/do-agent/build/do-agent_linux_amd64 /agent
RUN find /agent

CMD /agent/do-agent_linux_amd64
