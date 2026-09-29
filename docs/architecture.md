# Architecture

One executable, `cmd/imgd`, owns one HTTP server and its local caches.

Requests flow through the router, the shared skin lookup wrapper, and either
the raw skin handler or an image renderer. Skin lookup uses `mcclient` for
cached UUID, profile, and texture retrieval, backed by the Minecraft HTTP API.
There is no HTTP hop between skin lookup and rendering.

- `pkg/frontend`: embedded website and static files served at `/`.
- `pkg/imgd`: configuration, HTTP routing, startup, and graceful shutdown.
- `pkg/skin`: raw skins, downloads, cache checks, redirects, and ETag handling.
- `pkg/render`: rendered image routes and their processor registry.
- `pkg/mcskin`: image transformations and PNG/SVG encoding.
- `pkg/mcclient` and `pkg/minecraft`: cached lookups and upstream HTTP calls.
- `pkg/cache` and `pkg/storage`: existing cache formats and storage backends.
- `pkg/cache_converter`: legacy conversion, invoked as `imgd cacheconv`.

On SIGINT or SIGTERM, the HTTP server drains requests before closing its caches.
Startup errors return to the command instead of continuing with a nil server.

## Changes from the split services

Use `imgd` wherever `skind` or `processd` previously ran. Both sets of image
routes are served on the same HTTP port. Shared behavior is configured with
`imgd.*` flags and `IMGD_` environment variables. The remote `processd.skind-url`
setting is no longer needed.

The HTTP listen address, port, timeouts, instrumentation switch, application
flags, and cache settings are retained. The generic server framework and its
unused gRPC listener and flags have been removed. Framework-specific middleware,
request metrics, TLS, and logging flags are no longer provided; the application
retains its own cache, upstream, and image-request metrics. Existing metric names
containing `skind` or `processd` remain stable.

The separate Docker images, Compose monitoring stack, dashboard generation,
image publishing targets, archived application, and unused tiered-cache
experiment have been removed. Git history retains those files. The build now
uses ordinary `go build`, and DEFLATE compression uses one Go implementation
with the existing cache encoding.

The frontend is embedded at build time. Static files share the HTTP listener
with the image API and are served after the API routes. Unknown paths return
404, with no directory listing or catch-all homepage response.
