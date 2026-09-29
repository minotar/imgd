# Embedded Minotar frontend

The static website is adapted from https://github.com/minotar/minotar.net at
commit `0e699514641977c29399a1481124e180399c6e14`. Original site attribution
is retained in the page footer.

`static/` is embedded into imgd with `go:embed`; edit these files and rebuild
with `make`. There is no frontend build step or runtime asset directory.

Changes for serving from imgd:

- Images and the favicon use local routes.
- Example URLs use the current browser origin, including its port. Without
  JavaScript they remain usable relative paths.
- Tachyons 4.10.0 and Raleway 400/600/800 are bundled with license notices in
  `static/assets/`. Raleway was downloaded from Google Fonts' v37 URLs.
- Production analytics, canonical URL, and hosting attribution were removed.
- The IP.Board download points to the bundled original XML addon.

The handler serves existing files only, rejects directory listings, and does
not turn unknown image routes into homepage responses. Image API routes are
registered first. External project/plugin links are retained; loading the
page's styles, fonts, scripts, and images does not require a frontend CDN.
Minecraft lookups still use the configured upstream service.
