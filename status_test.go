package main

import (
	"github.com/op/go-logging"
	"runtime"
	"testing"
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
func TestStatusHandleMessageServed(t *testing.T) {
	stats.Served("test")
	runtime.Gosched()
	if stats.info.Served["test"] != 1 {
		t.Fatalf("Served[\"test\"] not 1, was %d", stats.info.Served["test"])
	}

	stats.Served("test")
	stats.Served("test")
	stats.Served("bacon")
	stats.Served("fromage")
	runtime.Gosched()
	if stats.info.Served["test"] != 3 {
		t.Fatalf("Served[\"test\"] not 3, was %d", stats.info.Served["test"])
	}
	if stats.info.Served["bacon"] != 1 {
		t.Fatalf("Served[\"bacon\"] not 1, was %d", stats.info.Served["bacon"])
	}
	if stats.info.Served["fromage"] != 1 {
		t.Fatalf("Served[\"fromage\"] not 1, was %d", stats.info.Served["fromage"])
	}
}
