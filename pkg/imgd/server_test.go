package imgd

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/util/log"
)

func diskConfig(t *testing.T) (Config, string) {
	t.Helper()
	var cfg Config
	flags := flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	cfg.RegisterFlags(flags)
	cfg.Logger = log.NewBuiltinLogger(4)
	dir := t.TempDir()
	for _, name := range []string{"uuid", "userdata", "textures"} {
		if err := flags.Set("cache."+name+".bolt-path", filepath.Join(dir, name+".db")); err != nil {
			t.Fatal(err)
		}
		if err := flags.Set("cache."+name+".name", t.Name()+name); err != nil {
			t.Fatal(err)
		}
	}
	cfg.Server.HTTPListenAddress = "127.0.0.1"
	cfg.Server.HTTPListenPort = 0
	return cfg, dir
}

func assertUnlocked(t *testing.T, path string) {
	t.Helper()
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("cache was not released: %v", err)
	}
	db.Close()
}

func TestStartupFailureClosesOpenedCaches(t *testing.T) {
	cfg, dir := diskConfig(t)
	cfg.McClient.CacheUserData.CacheType = "invalid"
	if app, err := New(cfg); err == nil || app != nil {
		t.Fatal("expected startup error")
	}
	assertUnlocked(t, filepath.Join(dir, "uuid.db"))
}

func TestRunContextClosesCaches(t *testing.T) {
	cfg, dir := diskConfig(t)
	cfg.McClient.CacheTextures.CacheType = "none"
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.RunContext(ctx); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"uuid", "userdata"} {
		assertUnlocked(t, filepath.Join(dir, name+".db"))
	}
}
