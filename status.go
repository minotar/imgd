package main

import (
	"encoding/json"
	"runtime"
	"time"
)

// The different MessageTypes for statusCollectorMessage
const (
	StatusTypeCacheHit = iota
	StatusTypeCacheMiss

	StatusTypeRequestServed
)

type statusCollectorMessage struct {
	// The type of message this is.
	MessageType uint

	// If MessageType==StatusTypeRequestServed, then the type of request that was served.
	RequestType string
}

type StatusCollector struct {
	info struct {
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
		// Number of skins in cache.
		CacheSize uint
		// Size of cache memory.
		CacheMem uint64
	}

	// Unix timestamp the process was booted at.
	StartedAt int64

	// Channel for feeding in input data.
	inputData chan statusCollectorMessage
}

func MakeStatsCollector() *StatusCollector {
	collector := &StatusCollector{}
	collector.StartedAt = time.Now().Unix()
	collector.info.Served = map[string]uint{}
	collector.inputData = make(chan statusCollectorMessage, 5)

	// Run a function every five seconds to collect time-based info.
	go func() {
		ticker := time.NewTicker(time.Second * 5)

		for {
			select {
			case <-ticker.C:
				collector.Collect()
			case msg := <-collector.inputData:
				collector.handleMessage(msg)
			}
		}
	}()

	return collector
}

// Message handler function, called inside goroutine.
func (s *StatusCollector) handleMessage(msg statusCollectorMessage) {
	switch msg.MessageType {
	case StatusTypeCacheHit:
		s.info.CacheHits += 1
	case StatusTypeCacheMiss:
		s.info.CacheMisses += 1
	case StatusTypeRequestServed:
		req := msg.RequestType
		if _, exists := s.info.Served[req]; exists {
			s.info.Served[req] += 1
		} else {
			s.info.Served[req] = 1
		}
	}
}

// Encodes the info struct to a JSON string byte slice
func (s *StatusCollector) ToJSON() []byte {
	results, _ := json.Marshal(s.info)
	return results
}

// "cron" function that updates current information
func (s *StatusCollector) Collect() {
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)

	s.info.Memory = memstats.Alloc
	s.info.Uptime = time.Now().Unix() - s.StartedAt
	s.info.CacheSize = cache.size()
	s.info.CacheMem = cache.memory()
}

// Increments the request counter for the specific type of request.
func (s *StatusCollector) Served(req string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeRequestServed,
		RequestType: req,
	}
}

// Should be called every time we serve a cached skin.
func (s *StatusCollector) HitCache() {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheHit,
	}
}

// Should be called every time we try and fail to serve a cached skin.
func (s *StatusCollector) MissCache() {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheMiss,
	}
}
