package main

import (
	"runtime"
	"testing"

	"github.com/op/go-logging"
)

func testSetupStatus() *StashingWriter {
	sw := new(StashingWriter)
	logBackend := logging.NewLogBackend(sw, "", 0)
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	setupCache()
	return sw
}

func TestStatusHandleMessageCacheHit(t *testing.T) {
	stats.HitCache()
	runtime.Gosched()
	if stats.info.CacheHits != 1 {
		t.Fatalf("CacheHits not 1, was %d", stats.info.CacheHits)
	}
}

func TestStatusHandleMessageCacheMiss(t *testing.T) {
	stats.MissCache()
	runtime.Gosched()
	if stats.info.CacheMisses != 1 {
		t.Fatalf("CacheMisses not 1, was %d", stats.info.CacheMisses)
	}
}

func TestStatusHandleMessageRequested(t *testing.T) {
	stats.Requested("test")
	runtime.Gosched()
	if stats.info.Requested["test"] != 1 {
		t.Fatalf("Requested[\"test\"] not 1, was %d", stats.info.Requested["test"])
	}

	stats.Requested("test")
	stats.Requested("test")
	stats.Requested("bacon")
	stats.Requested("fromage")
	runtime.Gosched()
	if stats.info.Requested["test"] != 3 {
		t.Fatalf("Requested[\"test\"] not 3, was %d", stats.info.Requested["test"])
	}
	if stats.info.Requested["bacon"] != 1 {
		t.Fatalf("Requested[\"bacon\"] not 1, was %d", stats.info.Requested["bacon"])
	}
	if stats.info.Requested["fromage"] != 1 {
		t.Fatalf("Requested[\"fromage\"] not 1, was %d", stats.info.Requested["fromage"])
	}
}

func TestStatusHandleMessageErrored(t *testing.T) {
	stats.Errored("test")
	runtime.Gosched()
	if stats.info.Errored["test"] != 1 {
		t.Fatalf("Errored[\"test\"] not 1, was %d", stats.info.Errored["test"])
	}

	stats.Errored("test")
	stats.Errored("test")
	stats.Errored("bacon")
	stats.Errored("fromage")
	runtime.Gosched()
	if stats.info.Errored["test"] != 3 {
		t.Fatalf("Errored[\"test\"] not 3, was %d", stats.info.Errored["test"])
	}
	if stats.info.Errored["bacon"] != 1 {
		t.Fatalf("Errored[\"bacon\"] not 1, was %d", stats.info.Errored["bacon"])
	}
	if stats.info.Errored["fromage"] != 1 {
		t.Fatalf("Errored[\"fromage\"] not 1, was %d", stats.info.Errored["fromage"])
	}
}
