# imgd

imgd serves the Minotar website, Minecraft skins, and avatars from a single Go application.

## Build and run

Install Go, then run:

```sh
make
IMGD_SERVER_HTTP_LISTEN_PORT=8080 ./cmd/imgd/imgd
```

Or use `go run ./cmd/imgd --server.http-listen-port=8080`.
Open `http://localhost:8080/` for the website and API examples. Its HTML, CSS,
fonts, JavaScript, favicon, and addon download are embedded in the executable.
Image examples use this same server.

The default port is 80. No C compiler, Node build, containers, or separate web
server is required.

Use `./cmd/imgd/imgd --help` for configuration. Flags also accept environment
variables prefixed with `IMGD_`, uppercasing and replacing `-` and `.` with `_`.
For example, `--imgd.redirect-username=false` is
`IMGD_IMGD_REDIRECT_USERNAME=false`.

## Docker

Build and run with a named volume to preserve the caches:

```sh
docker build -t imgd .
docker run --rm --name imgd -p 8080:8080 -v imgd-data:/data imgd
```

Open `http://localhost:8080/`. The image runs as a non-root user (UID/GID 10001),
listens on port 8080, and stores its three Bolt caches in `/data`. If you bind
mount a host directory instead, make it writable by UID 10001.

Pass configuration through Docker's `-e` option or append flags to the image
name, for example:

```sh
docker run --rm -p 8080:8080 -v imgd-data:/data imgd --imgd.redirect-username=false
```

Stop the running container before reusing its cache volume in another container
or with `imgd cacheconv`.

## Frontend

The site lives in `pkg/frontend/static/`. Edit its files and run `make` to rebuild
the embedded assets. The executable can run from any working directory.
See [frontend source notes](pkg/frontend/README.md) for upstream attribution.

## HTTP API

All image routes accept a Minecraft username or UUID. Dashed UUIDs redirect to
plain UUIDs. Usernames redirect to UUIDs by default; disable this with
`--imgd.redirect-username=false` to return images directly.

| Route | Output |
| --- | --- |
| `/skin/<user>[.png]` | Original skin |
| `/download/<user>[.png]` | Skin attachment |
| `/avatar/<user>[/<width>][.png\|.svg]` | Head |
| `/helm/<user>[/<width>][.png\|.svg]` | Head with helmet |
| `/cube/<user>[/<width>][.png\|.svg]` | Isometric head |
| `/cubehelm/<user>[/<width>][.png\|.svg]` | Isometric head with helmet |
| `/bust/<user>[/<width>][.png\|.svg]` | Upper body |
| `/body/<user>[/<width>][.png\|.svg]` | Full body |
| `/armor/bust/<user>[/<width>][.png\|.svg]` | Upper body with outer layer |
| `/armor/body/<user>[/<width>][.png\|.svg]` | Full body with outer layer |

`armour` is also accepted in place of `armor`. Rendered images default to PNG
at 180 pixels wide; requested widths are clamped to 8–300 pixels.
Missing skins use the built-in Steve skin. ETags, cache-control headers, and
permissive CORS remain configurable.

Operational endpoints are `/healthcheck`, `/dbsize`, `/metrics`,
`/debug/pprof/`, and `/debug/fgprof`. Disable metrics and pprof with
`--server.register-instrumentation=false`.

## Caches

UUIDs, user data, and textures use local Bolt files by default:
`/tmp/bolt_cache_uuid.db`, `/tmp/bolt_cache_userdata.db`, and
`/tmp/bolt_cache_textures.db`. Configure persistent locations with
`--cache.uuid.bolt-path`, `--cache.userdata.bolt-path`, and
`--cache.textures.bolt-path`. Existing cache formats and the Badger and
migration backends are retained. Texture caching can be disabled with
`--cache.textures.backend=none`; UUID and user-data caches are required.

The old Redis/Bolt conversion utility is available in the same executable:

```sh
./cmd/imgd/imgd cacheconv --help
```

Its existing flags and `CACHECONV_` environment variables are unchanged.
Run conversions with the application stopped so cache files are not shared.

## Development

```sh
make test
```

All packages, including the Minecraft client, belong to one Go module and are
covered by `go test ./...`. Tests use local Minecraft fixtures.
See [architecture](docs/architecture.md) for the package layout and migration notes.

## Thanks

Thanks to [lukegb](https://github.com/lukegb) for porting the original PHP script to Go.
