package mcclient

import (
	"io"
	"net/http"
	"testing"

	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/cache/lru_cache"
	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/minecraft/mockminecraft"
	"github.com/minotar/imgd/pkg/util/log"
)

type testReporter interface {
	Fatalf(format string, args ...interface{})
}

func newMcClient(t testReporter, size int) (*McClient, func()) {
	logger := log.NewBuiltinLogger(1)
	lruCache, err := lru_cache.NewLruCache(lru_cache.NewLruCacheConfig(size,
		cache.CacheConfig{
			Name:   "LruCache",
			Logger: logger,
		},
	))

	if err != nil {
		t.Fatalf("Error creating LruCache: %s", err)
	}

	mux := mockminecraft.ReturnMux()
	rt, shutdown := mockminecraft.Setup(mux)

	minecraftClient := minecraft.Minecraft{
		Client: &http.Client{
			Transport: rt,
		},
		Cfg: minecraft.Config{
			UUIDAPIConfig: minecraft.UUIDAPIConfig{
				SessionServerURL: "http://example.com/session/minecraft/profile/",
				ProfileURL:       "http://example.com/users/profiles/minecraft/",
			},
		},
	}

	mcClient := &McClient{
		API: &minecraftClient,
	}
	mcClient.Caches.UUID = lruCache
	mcClient.Caches.UserData = lruCache
	mcClient.Caches.Textures = lruCache

	return mcClient, shutdown
}

func TestUsername(t *testing.T) {
	logger := log.NewBuiltinLogger(1)
	mcClient, shutdown := newMcClient(t, 5)
	defer shutdown()

	uuidEntry, err := mcClient.GetUUIDEntry(logger, "lukehandle")
	if err != nil {
		t.Fatalf("Get UUID ENtry failed: %v", err)
	}

	if uuidEntry.UUID != "5c115ca73efd41178213a0aff8ef11e0" {
		t.Errorf("UUID was not expected: %v", uuidEntry)
	}
}

func BenchmarkSkinCacheHit(b *testing.B) {
	logger := log.NewBuiltinLogger(1)
	mcClient, shutdown := newMcClient(b, 5)
	defer shutdown()
	mcClient.GetSkinFromReq(logger, UserReq{Username: "clone1018"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userReq := UserReq{Username: "clone1018"}
		skin := mcClient.GetSkinFromReq(logger, userReq)
		if skin.Hash != "a04a26d10218668a632e419ab073cf57" {
			b.Fatalf("Skin hash was not as expected: %v", skin)
		}
	}
}

func BenchmarkSkinCacheMiss(b *testing.B) {
	logger := log.NewBuiltinLogger(1)
	// Cache size of 1 means we'll be constantly inserting/evicting
	mcClient, shutdown := newMcClient(b, 1)
	defer shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userReq := UserReq{Username: "clone1018"}
		skin := mcClient.GetSkinFromReq(logger, userReq)
		if skin.Hash != "a04a26d10218668a632e419ab073cf57" {
			b.Fatalf("Skin hash was not as expected: %v", skin)
		}
	}
}

func BenchmarkSkinBufCacheHit(b *testing.B) {
	logger := log.NewBuiltinLogger(1)
	mcClient, shutdown := newMcClient(b, 5)
	defer shutdown()
	mcClient.GetSkinFromReq(logger, UserReq{Username: "clone1018"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userReq := UserReq{Username: "clone1018"}
		_, textureIO := mcClient.GetSkinBufferFromReq(logger, userReq)
		bytes, err := io.ReadAll(textureIO)
		if err != nil {
			b.Fatalf("oops")
		}
		if len(bytes) != 1544 {
			b.Fatal("Bytes were too short")
		}
	}
}

func BenchmarkSkinBufCacheMiss(b *testing.B) {
	logger := log.NewBuiltinLogger(1)
	// Cache size of 1 means we'll be constantly inserting/evicting
	mcClient, shutdown := newMcClient(b, 1)
	defer shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userReq := UserReq{Username: "clone1018"}
		_, textureIO := mcClient.GetSkinBufferFromReq(logger, userReq)
		bytes, err := io.ReadAll(textureIO)
		if err != nil {
			b.Fatalf("oops")
		}
		if len(bytes) != 1544 {
			b.Fatal("Bytes were too short")
		}
	}
}
