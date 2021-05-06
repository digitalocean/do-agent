FROM golang:1.16.3 as build
ENV DOCKER_BUILD=1
ADD . /home/do-agent
WORKDIR /home/do-agent
RUN set -x && \
        make clean build

FROM alpine as alpine
RUN mkdir -p /host
RUN set -x && \
        apk add --no-cache dumb-init

FROM gcr.io/distroless/base
COPY --from=alpine /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=alpine /host /host
COPY --from=build /home/do-agent/target/do-agent-linux-amd64 /bin/do-agent
VOLUME /host/proc /host/sys
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/do-agent", "--path.procfs", "/host/proc", "--path.sysfs", "/host/sys"]
