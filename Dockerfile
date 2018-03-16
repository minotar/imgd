#build stage
FROM golang:alpine AS builder
WORKDIR /go/src/github.com/minotar/imgd
COPY . .
RUN apk add --no-cache git
RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/imgd /imgd
COPY config.example.gcfg /config.gcfg
ENTRYPOINT ./imgd
LABEL Name=imgd Version=2.11.0
EXPOSE 8000
