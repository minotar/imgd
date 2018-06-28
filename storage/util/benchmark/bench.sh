#!/bin/bash

container=$(docker run -d --rm -v redis.conf:/usr/local/etc/redis/redis.conf -p 6379:6379 redis redis-server /usr/local/etc/redis/redis.conf)
go test -benchmem -bench ^BenchmarkR
docker rm -f "${container}"
