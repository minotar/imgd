package main

import (
	"testing"
	"time"

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
	stats = MakeStatsCollector()
	stats.CacheUUID("hit")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.CacheUUID.Hits != 1 {
		t.Fatalf("CacheUUID.Hits not 1, was %d", stats.info.CacheUUID.Hits)
	}
}

func TestStatusHandleMessageCacheFresh(t *testing.T) {
	stats = MakeStatsCollector()
	stats.CacheUUID("fresh")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.CacheUUID.FreshHits != 1 {
		t.Fatalf("CacheUUID.FreshHits not 1, was %d", stats.info.CacheUUID.FreshHits)
	}
}

func TestStatusHandleMessageCacheStale(t *testing.T) {
	stats = MakeStatsCollector()
	stats.CacheUUID("stale")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.CacheUUID.StaleHits != 1 {
		t.Fatalf("CacheUUID.StaleHits not 1, was %d", stats.info.CacheUUID.StaleHits)
	}
}

func TestStatusHandleMessageCacheMiss(t *testing.T) {
	stats = MakeStatsCollector()
	stats.CacheUUID("miss")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.CacheUUID.Misses != 1 {
		t.Fatalf("CacheUUID.Misses not 1, was %d", stats.info.CacheUUID.Misses)
	}
}

func TestStatusHandleMessageRequested(t *testing.T) {
	stats = MakeStatsCollector()
	stats.Requested("test")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.Requested["test"] != 1 {
		t.Fatalf("Requested[\"test\"] not 1, was %d", stats.info.Requested["test"])
	}

	stats.Requested("test")
	stats.Requested("test")
	stats.Requested("bacon")
	stats.Requested("fromage")
	time.Sleep(time.Duration(1) * time.Millisecond)
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

func TestStatusHandleMessageUserRequested(t *testing.T) {
	stats = MakeStatsCollector()
	stats.UserRequested("test")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.UserRequested["test"] != 1 {
		t.Fatalf("UserRequested[\"test\"] not 1, was %d", stats.info.UserRequested["test"])
	}

	stats.UserRequested("test")
	stats.UserRequested("test")
	stats.UserRequested("bacon")
	stats.UserRequested("fromage")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.UserRequested["test"] != 3 {
		t.Fatalf("UserRequested[\"test\"] not 3, was %d", stats.info.UserRequested["test"])
	}
	if stats.info.UserRequested["bacon"] != 1 {
		t.Fatalf("UserRequested[\"bacon\"] not 1, was %d", stats.info.UserRequested["bacon"])
	}
	if stats.info.UserRequested["fromage"] != 1 {
		t.Fatalf("UserRequested[\"fromage\"] not 1, was %d", stats.info.UserRequested["fromage"])
	}
}

func TestStatusHandleMessageErrored(t *testing.T) {
	stats = MakeStatsCollector()
	stats.Errored("test")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if stats.info.Errored["test"] != 1 {
		t.Fatalf("Errored[\"test\"] not 1, was %d", stats.info.Errored["test"])
	}

	stats.Errored("test")
	stats.Errored("test")
	stats.Errored("bacon")
	stats.Errored("fromage")
	time.Sleep(time.Duration(1) * time.Millisecond)
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
