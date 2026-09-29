FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/imgd ./cmd/imgd

FROM alpine:3.23

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 imgd \
    && adduser -S -D -H -u 10001 -G imgd imgd \
    && mkdir /data \
    && chown imgd:imgd /data

COPY --from=build /out/imgd /usr/local/bin/imgd

ENV IMGD_SERVER_HTTP_LISTEN_PORT=8080 \
    IMGD_CACHE_UUID_BOLT_PATH=/data/uuid.db \
    IMGD_CACHE_USERDATA_BOLT_PATH=/data/userdata.db \
    IMGD_CACHE_TEXTURES_BOLT_PATH=/data/textures.db

USER imgd:imgd
WORKDIR /data
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/imgd"]
