# imgd

imgd is a simple avatar serving system. You're probably looking for [Minotar.net](https://github.com/minotar/minotar.net)

`imgd` is a single binary for quickly getting up and running with skins + processed avatars.

We also provide `skind` (just skins) and `processd` (just processing). You should then configure `processd` to point at your instance(s) or `skind`.

## How to install?
Installation is simple - however it requires an installation of [Go](http://golang.org). Ensure you are set up there before trying these commands.


```bash
git clone https://github.com/minotar/imgd
cd imgd

make imgd
```

This should compile a binary to `./cmd/imgd/imgd` which can then be used for serving your images.

Try `./cmd/imgd/imgd --help` to see the options. You can also pass ENV vars (uppercasing and replacing `-`/`.` with `_`).

```bash

IMGD_SERVER_HTTP_LISTEN_PORT=8080 ./cmd/imgd/imgd
```

## Thanks
Big thanks to [lukegb](https://github.com/lukegb) for porting the old version of this script from PHP to Go.
