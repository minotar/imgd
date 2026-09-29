package imgd

import (
	"bytes"
	"flag"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/lru_cache"
	"github.com/minotar/imgd/pkg/mcclient"
	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/minecraft/mockminecraft"
	"github.com/minotar/imgd/pkg/render"
	"github.com/minotar/imgd/pkg/util/log"
)

const testUUID = "d9135e082f2244c89cb0bee234155292"

func testApp(t *testing.T) *Imgd {
	t.Helper()
	var cfg Config
	cfg.RegisterFlags(flag.NewFlagSet(t.Name(), flag.ContinueOnError))
	cfg.Logger = log.NewBuiltinLogger(4)
	rt, shutdown := mockminecraft.Setup(mockminecraft.ReturnMux())
	t.Cleanup(shutdown)
	mc := mcclient.NewMcClient(&cfg.McClient)
	mc.API.Client = &http.Client{Transport: rt}
	c, err := lru_cache.NewLruCache(lru_cache.NewLruCacheConfig(100, cache.CacheConfig{Name: t.Name(), Logger: cfg.Logger}))
	if err != nil {
		t.Fatal(err)
	}
	mc.Caches.UUID, mc.Caches.UserData, mc.Caches.Textures = c, c, c
	return &Imgd{Cfg: cfg, McClient: mc, ProcessRoutes: render.DefaultProcessRoutes}
}

func request(h http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestImageEndpoints(t *testing.T) {
	app := testApp(t)
	app.Cfg.RedirectUsername = false
	handler := app.Handler()
	for _, resource := range []string{"avatar", "helm", "cube", "cubehelm", "bust", "body", "armor/bust", "armour/bust", "armor/body", "armour/body"} {
		for _, user := range []string{testUUID, "clone1018"} {
			for _, suffix := range []string{"", ".png", ".svg", "/64", "/64.png", "/64.svg"} {
				path := "/" + resource + "/" + user + suffix
				t.Run(path, func(t *testing.T) {
					w := request(handler, path)
					if w.Code != http.StatusOK {
						t.Fatalf("status %d: %s", w.Code, w.Body.String())
					}
					if w.Header().Get("Access-Control-Allow-Origin") != "*" {
						t.Error("missing CORS header")
					}
					if w.Header().Get("Cache-Control") != "public, max-age=21600" {
						t.Error("cache TTL changed")
					}
					if w.Header().Get("ETag") == "" {
						t.Error("missing ETag")
					}
					if strings.HasSuffix(suffix, ".svg") {
						if w.Header().Get("Content-Type") != "image/svg+xml" || !strings.Contains(w.Body.String(), "<svg") {
							t.Fatal("invalid SVG response")
						}
					} else {
						image, err := png.Decode(bytes.NewReader(w.Body.Bytes()))
						if err != nil {
							t.Fatal(err)
						}
						width := 180
						if strings.HasPrefix(suffix, "/64") {
							width = 64
						}
						if image.Bounds().Dx() != width {
							t.Errorf("width = %d, want %d", image.Bounds().Dx(), width)
						}
					}
				})
			}
		}
	}
	for _, resource := range []string{"skin", "download"} {
		for _, user := range []string{testUUID, "clone1018"} {
			for _, suffix := range []string{"", ".png"} {
				w := request(handler, "/"+resource+"/"+user+suffix)
				if w.Code != http.StatusOK {
					t.Fatalf("%s: status %d", resource, w.Code)
				}
				if _, err := png.Decode(w.Body); err != nil {
					t.Fatal(err)
				}
				if resource == "download" && !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment; filename=\"") {
					t.Error("missing download filename")
				}
			}
		}
	}
}

func TestRedirectsAndConditionalRequests(t *testing.T) {
	app := testApp(t)
	handler := app.Handler()
	for _, resource := range []string{"skin", "download", "avatar", "helm", "cube", "cubehelm", "bust", "body", "armor/bust", "armour/bust", "armor/body", "armour/body"} {
		w := request(handler, "/"+resource+"/clone1018")
		if w.Code != http.StatusFound || w.Header().Get("Location") != "/"+resource+"/"+testUUID {
			t.Fatalf("%s: unexpected redirect: %d %s", resource, w.Code, w.Header().Get("Location"))
		}
		w = request(handler, "/"+resource+"/d9135e08-2f22-44c8-9cb0-bee234155292")
		if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "/"+resource+"/"+testUUID {
			t.Fatalf("%s: dashed UUID redirect changed", resource)
		}
		path := "/" + resource + "/" + testUUID
		w = request(handler, path)
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("If-None-Match", w.Header().Get("ETag"))
		conditional := httptest.NewRecorder()
		handler.ServeHTTP(conditional, req)
		if conditional.Code != http.StatusNotModified || conditional.Body.Len() != 0 {
			t.Fatalf("%s: conditional response = %d", resource, conditional.Code)
		}
	}
}

func TestFallbackAndOperationalEndpoints(t *testing.T) {
	app := testApp(t)
	app.Cfg.RedirectUsername = false
	app.Cfg.UseETags = false
	app.Cfg.CorsAllowAll = false
	app.Cfg.CacheControlTTL = time.Minute
	handler := app.Handler()
	for _, user := range []string{"missinguser", "500session", "notexture", minecraft.SteveUUID} {
		w := request(handler, "/avatar/"+user+"/32.png")
		if w.Code != http.StatusOK {
			t.Fatalf("fallback for %s: %d", user, w.Code)
		}
		if _, err := png.Decode(w.Body); err != nil {
			t.Fatal(err)
		}
		if w.Header().Get("ETag") != "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("disabled headers are still present")
		}
		if w.Header().Get("Cache-Control") != "public, max-age=60" {
			t.Error("configured TTL not applied")
		}
	}
	for _, path := range []string{"/", "/assets/tachyons.min.css", "/favicon.ico", "/robots.txt", "/addons/ipb/minotar.xml", "/healthcheck", "/dbsize", "/metrics", "/debug/pprof/"} {
		if w := request(handler, path); w.Code != http.StatusOK {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	if w := request(handler, "/not-a-route"); w.Code != http.StatusNotFound {
		t.Errorf("unknown route: %d", w.Code)
	}
	app.Cfg.Server.RegisterInstrumentation = false
	if w := request(app.Handler(), "/metrics"); w.Code != http.StatusNotFound {
		t.Error("metrics should be disabled")
	}
}

func TestUsernameRedirectPreservesImageOptions(t *testing.T) {
	handler := testApp(t).Handler()
	for _, path := range []string{
		"/avatar/clone1018/64", "/avatar/clone1018/64.png", "/avatar/clone1018/64.svg",
		"/helm/clone1018.svg", "/armor/body/clone1018/100.png", "/skin/clone1018.png",
		"/download/clone1018.png", "/avatar/clone1018/64?example=true",
	} {
		response := request(handler, path)
		expected := strings.Replace(path, "clone1018", testUUID, 1)
		if response.Code != http.StatusFound || response.Header().Get("Location") != expected {
			t.Errorf("%s: got %d %s, want 302 %s", path, response.Code, response.Header().Get("Location"), expected)
		}
		image := request(handler, expected)
		if image.Code != http.StatusOK {
			t.Errorf("redirect destination %s: %d", expected, image.Code)
		}
	}
}
