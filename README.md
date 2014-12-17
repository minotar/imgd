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
$ ./imgd
```
There you have it! Go visit your installation at *your-ip*:8000 to view it in action. If you wish to change the address the server listens on, you can do so by editing `config.gcfg` (it's like an `ini` file).

## Thanks
Big thanks to [lukegb](https://github.com/lukegb) for porting the old version of this script from PHP to Go.
