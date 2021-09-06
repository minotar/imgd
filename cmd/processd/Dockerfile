FROM golang:1.16.7 as build

COPY . /src/processd
WORKDIR /src/processd
RUN make clean && make processd

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=build /src/processd/cmd/processd/processd /usr/bin/processd

RUN addgroup -g 10001 -S processd && \
    adduser -u 10001 -S processd -G processd

RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

USER processd
EXPOSE 8080
ENTRYPOINT [ "/usr/bin/processd" ]
CMD [ \
    "--server.http-listen-port=8080", \
    "--processd.skind-url=http://skind:4643/skin/" \
    ]
