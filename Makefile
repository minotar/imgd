GO ?= go

.PHONY: all imgd test clean

all: imgd

imgd:
	$(GO) build -o cmd/imgd/imgd ./cmd/imgd

test:
	$(GO) test ./...

clean:
	rm -f cmd/imgd/imgd cmd/imgd/imgd.exe
