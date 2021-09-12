# Heavily influenced by https://github.com/grafana/loki/blob/806d6a503e1746a158f332d055f5d9fbd8b28d1a/Makefile
# Original is Licensed Apache 2.0


MOD_FLAG=


IMAGE_PREFIX ?= minotar

IMAGE_TAG := $(shell ./build/image-tag)


# Version info for binaries
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# We don't want find to scan inside a bunch of directories, to accelerate the
# 'make: Entering directory '/src/loki' phase.
DONT_FIND := -name examples -prune -o -name legacy -prune -o -name .git -prune -o -name .cache -prune -o -name .pkg -prune -o

# These are all the application files, they are included in the various binary rules as dependencies
# to make sure binaries are rebuilt if any source files change.
APP_GO_FILES := $(shell find . $(DONT_FIND) -name .y.go -prune -o -name .pb.go -prune -o -name cmd -prune -o -type f -name '*.go' -print)



# Build flags
VPREFIX := github.com/minotar/imgd/pkg/build
GO_CGO       ?= 1
GO_LDFLAGS   := -X $(VPREFIX).Branch=$(GIT_BRANCH) -X $(VPREFIX).Version=$(IMAGE_TAG) -X $(VPREFIX).Revision=$(GIT_REVISION) -X $(VPREFIX).BuildUser=$(shell whoami)@$(shell hostname) -X $(VPREFIX).BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_FLAGS     := -ldflags "-extldflags \"-static\" -s -w $(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)
DYN_GO_FLAGS := -ldflags "-s -w $(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)
# Per some websites I've seen to add `-gcflags "all=-N -l"`, the gcflags seem poorly if at all documented
# the best I could dig up is -N disables optimizations and -l disables inlining which should make debugging match source better.
# Also remove the -s and -w flags present in the normal build which strip the symbol table and the DWARF symbol table.
DEBUG_GO_FLAGS     := -gcflags "all=-N -l" -ldflags "-extldflags \"-static\" $(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)
DYN_DEBUG_GO_FLAGS := -gcflags "all=-N -l" -ldflags "$(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)


# Protobuf files
PROTO_DEFS := $(shell find . $(DONT_FIND) -type f -name '*.proto' -print)
PROTO_GOS := $(patsubst %.proto,%.pb.go,$(PROTO_DEFS))



all: skind processd imgd cacheconv



skind: protos cmd/skind/skind
skind-debug: protos cmd/skind/skind-debug

cmd/skind/skind: $(APP_GO_FILES) cmd/skind/main.go
	CGO_ENABLED=$(GO_CGO) go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

cmd/skind/skind-debug: $(APP_GO_FILES) cmd/skind/main.go
	CGO_ENABLED=$(GO_CGO) go build $(DEBUG_GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)



processd: cmd/processd/processd
processd-debug: cmd/processd/processd-debug

cmd/processd/processd: $(APP_GO_FILES) cmd/processd/main.go
	CGO_ENABLED=$(GO_CGO) go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

cmd/processd/processd-debug: $(APP_GO_FILES) cmd/processd/main.go
	CGO_ENABLED=$(GO_CGO) go build $(DEBUG_GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)



imgd: cmd/imgd/imgd
imgd-debug: cmd/imgd/imgd-debug

cmd/imgd/imgd: $(APP_GO_FILES) cmd/imgd/main.go
	CGO_ENABLED=$(GO_CGO) go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

cmd/imgd/imgd-debug: $(APP_GO_FILES) cmd/imgd/main.go
	CGO_ENABLED=$(GO_CGO) go build $(DEBUG_GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)



cacheconv: cmd/cacheconv/cacheconv
cacheconv-debug: cmd/cacheconv/cacheconv-debug

cmd/cacheconv/cacheconv: $(APP_GO_FILES) cmd/cacheconv/main.go
	CGO_ENABLED=$(GO_CGO) go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

cmd/cacheconv/cacheconv-debug: $(APP_GO_FILES) cmd/cacheconv/main.go
	CGO_ENABLED=$(GO_CGO) go build $(DEBUG_GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)



clean:
	rm -rf cmd/skind/skind
	rm -rf cmd/skind/skind-debug
	rm -rf cmd/processd/processd
	rm -rf cmd/processd/processd-debug
	rm -rf cmd/imgd/imgd
	rm -rf cmd/imgd/imgd-debug
	rm -rf cmd/cacheconv/cacheconv
	rm -rf cmd/cacheconv/cacheconv-debug


#############
# Protobufs #
#############

protos: $(PROTO_GOS)

# use with care. This signals to make that the proto definitions don't need recompiling.
touch-protos:
	for proto in $(PROTO_GOS); do [ -f "./$${proto}" ] && touch "$${proto}" && echo "touched $${proto}"; done

%.pb.go: $(PROTO_DEFS)
	protoc --proto_path=./ --go_out=${GOPATH}/src ./$(patsubst %.pb.go,%.proto,$@)


#################
# Docker Images #
#################

images: skind-image processd-image imgd-image
images-push: skind-image-push processd-image-push imgd-image-push


# skind
skind-image:
	docker build -t $(IMAGE_PREFIX)/skind:$(IMAGE_TAG) -f cmd/skind/Dockerfile .

skind-image-push:
	docker push $(IMAGE_PREFIX)/skind:$(IMAGE_TAG)

# processd
processd-image:
	docker build -t $(IMAGE_PREFIX)/processd:$(IMAGE_TAG) -f cmd/processd/Dockerfile .

processd-image-push:
	docker push $(IMAGE_PREFIX)/processd:$(IMAGE_TAG)

# imgd
imgd-image:
	docker build -t $(IMAGE_PREFIX)/imgd:$(IMAGE_TAG) -f cmd/imgd/Dockerfile .

imgd-image-push:
	docker push $(IMAGE_PREFIX)/imgd:$(IMAGE_TAG)

