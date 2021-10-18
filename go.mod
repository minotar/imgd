module github.com/minotar/imgd

go 1.16

require (
	github.com/4kills/go-libdeflate v1.3.1
	github.com/ajstarks/svgo v0.0.0-20210406150507-75cfd577ce75
	github.com/boltdb/bolt v1.3.1
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/disintegration/gift v1.2.1
	github.com/disintegration/imaging v1.6.2
	github.com/felixge/fgprof v0.9.1
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/kamaln7/envy v1.0.1-0.20200811133559-2c7680e4c27d
	github.com/klauspost/compress v1.13.6
	github.com/levenlabs/golib v0.0.0-20180911183212-0f8974794783 // indirect
	github.com/mediocregopher/radix.v2 v0.0.0-20181115013041-b67df6e626f9
	github.com/minotar/imgd/pkg/minecraft v0.0.0-00010101000000-000000000000
	github.com/oschwald/geoip2-golang v1.5.0 // indirect
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.23.0
	github.com/spf13/pflag v1.0.5
	github.com/weaveworks/common v0.0.0-20210506120931-f2676019da11
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.18.1
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e // indirect
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a // indirect
	google.golang.org/protobuf v1.23.0
)

replace github.com/minotar/imgd/pkg/minecraft => ./pkg/minecraft
