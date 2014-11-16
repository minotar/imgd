# imgd

imgd is a simple avatar serving system. You're probably looking for [Minotar.net](https://github.com/minotar/minotar.net)

## How to install?
Installation is simple - however it requires an installation of [Go](http://golang.org). Follow the instructions below for a comprehensive, step by step installation.
```bash
$ git clone https://github.com/minotar/imgd
$ cd imgd

$ export GOPATH=`pwd`
$ go get

$ go build
```
After you run `go build`, golang should automatically generate you an executable file (named `imgd`). Executing the file is simple: simply run:
```bash
$ export IMGD_LISTENON=:8000
$ ./imgd
```
There you have it! Go visit your installation at *your-ip*:8000 to view it in action.

## Understanding Headers
We use a couple of headers to help in understanding how something is served, here they are:

x-requested

- returns: processed
- explain: if Minotar processed your avatar

x-result:

- returns: ok,failed
- explain: ok on successful GET from s3, failed on failed GET from s3

x-timing:

- returns: fetch time, process time, resize time, whole process
- example: 48+0+4=52

## Thanks
Big thanks to [lukegb](https://github.com/lukegb) for porting the old version of this script from PHP to Go.
