FROM golang:1.16.7 as build

COPY . /src/imgd
WORKDIR /src/imgd
RUN make clean && make imgd

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=build /src/imgd/cmd/imgd/imgd /usr/bin/imgd

RUN addgroup -g 10001 -S imgd && \
    adduser -u 10001 -S imgd -G imgd
RUN mkdir -p /imgd/ && \
    chown -R imgd:imgd /imgd

RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

USER imgd
EXPOSE 8080
ENTRYPOINT [ "/usr/bin/imgd" ]
CMD [ \
    "--server.http-listen-port=8080", \
    "--cache.uuid.bolt-path=/imgd/bolt_cache_uuid.db", \
    "--cache.userdata.bolt-path=/imgd/bolt_cache_usertdata.db", \
    "--cache.textures.bolt-path=/imgd/bolt_cache_textures.db" \
    ]
