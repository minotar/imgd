package main

import (
	"encoding/json"
	"runtime"
	"time"
)

type StatusCollector struct {
	Info struct {
		// Number of bytes allocated to the process.
		Memory uint64
		// Time in seconds the process has been running for
		Uptime int64
		// Number of times a request type has been served.
		Served map[string]uint
		// Number of times skins have been served from the cache.
		CacheHits uint
		// Number of times skins have failed to be served from the cache.
		CacheMisses uint
		// Size of cached skins
		Cached uint64
	}

	// Unix timestamp the process was booted at.
	StartedAt int64
}

func MakeStatsCollector() *StatusCollector {
	collector := &StatusCollector{}
	collector.StartedAt = time.Now().Unix()
	collector.Info.Served = map[string]uint{}

	// Run a function every five seconds to collect time-based info.
	go func() {
		for {
			time.Sleep(time.Second * 5)
			collector.Collect()
		}
	}()

	return collector
}

// Encodes the info struct to a JSON string byte slice
func (s *StatusCollector) ToJSON() []byte {
	results, _ := json.Marshal(s.Info)
	return results
}

// "cron" function that updates current information
func (s *StatusCollector) Collect() {
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)

	s.Info.Memory = memstats.Alloc
	s.Info.Uptime = time.Now().Unix() - s.StartedAt
	s.Info.Cached = cache.memory()
}

// Increments the request counter for the specific type of request.
func (s *StatusCollector) Served(req string) {
	if _, exists := s.Info.Served[req]; exists {
		s.Info.Served[req] += 1
	} else {
		s.Info.Served[req] = 1
	}
}

// Should be called every time we serve a cached skin.
func (s *StatusCollector) HitCache() {
	s.Info.CacheHits += 1
}

// Should be called every time we try and fail to serve a cached skin.
func (s *StatusCollector) MissCache() {
	s.Info.CacheMisses += 1
}
