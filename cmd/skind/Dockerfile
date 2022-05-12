FROM golang:1.17.2 as build

WORKDIR /src/minotar

COPY go.mod .
COPY go.sum .
COPY pkg/minecraft/go.mod pkg/minecraft/go.mod
RUN go mod download -x

COPY . /src/minotar
RUN make clean && make skind

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=build /src/minotar/cmd/skind/skind /usr/bin/skind

RUN addgroup -g 10001 -S skind && \
    adduser -u 10001 -S skind -G skind
RUN mkdir -p /skind/ && \
    chown -R skind:skind /skind

RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

USER skind
EXPOSE 4643
ENTRYPOINT [ "/usr/bin/skind" ]
CMD [ \
    "--server.http-listen-port=4643", \
    "--cache.uuid.bolt-path=/skind/bolt_cache_uuid.db", \
    "--cache.userdata.bolt-path=/skind/bolt_cache_usertdata.db", \
    "--cache.textures.bolt-path=/skind/bolt_cache_textures.db" \
    ]
