FROM alpine as alpine
RUN mkdir -p /host
RUN set -x && \
        apk add --no-cache dumb-init

FROM gcr.io/distroless/base
COPY --from=alpine /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=alpine /host /host
VOLUME /host/proc /host/sys
ADD target/do-agent-linux-amd64 /bin/do-agent
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/do-agent", "--path.procfs", "/host/proc", "--path.sysfs", "/host/sys"]
