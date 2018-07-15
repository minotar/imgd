// Todo: Could we instead use the prometheus.DefaultGatherer and Gather() to then use the Prometheus counters vs. duplicating it
package main

import (
	"encoding/json"
	"runtime"
	"time"
)

// The different MessageTypes for statusCollectorMessage
const (
	StatusTypeErrored = iota

	StatusTypeRequested
	StatusTypeAPIRequested

	StatusTypeCacheUUID
	StatusTypeCacheUserData
	StatusTypeCacheSkin
	StatusTypeCacheSkinTransient
)

type statusCollectorMessage struct {
	// The type of message this is.
	MessageType uint

	// If MessageType == StatusTypeRequested, StatusTypeAPIRequested or StatusTypeErrored then this is the state we are reporting.
	StatusType string
}

type cacheStats struct {
	Hits      uint
	FreshHits uint
	StaleHits uint
	Misses    uint
	Errors    uint
	Length    uint
	Size      uint64
}

type StatusCollector struct {
	info struct {
		// Number of bytes allocated to the process.
		ImgdMem uint64
		// Time in seconds the process has been running for
		Uptime int64
		// Number of times an error has been recorded.
		Errored map[string]uint
		// Number of times a request type has been requested.
		Requested map[string]uint
		// Number of times an API request type has been made.
		APIRequested map[string]uint
		// Cache stats for the Username->UUID cache
		CacheUUID cacheStats
		// Cache stats for the UUID->UserData cache
		CacheUserData cacheStats
		// Cache stats for the SkinPath->Skin cache
		CacheSkin cacheStats
		// Cache stats for the SkinPath->SkinError cache
		CacheSkinTransient cacheStats
	}

	// Unix timestamp the process was booted at.
	StartedAt int64

	// Channel for feeding in input data.
	inputData chan statusCollectorMessage
}

func MakeStatsCollector() *StatusCollector {
	collector := &StatusCollector{}
	collector.StartedAt = time.Now().Unix()
	collector.info.Errored = map[string]uint{}
	collector.info.Requested = map[string]uint{}
	collector.info.APIRequested = map[string]uint{}
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
	case StatusTypeErrored:
		err := msg.StatusType
		errorCounter.WithLabelValues(err).Inc()
		if _, exists := s.info.Errored[err]; exists {
			s.info.Errored[err]++
		} else {
			s.info.Errored[err] = 1
		}
	case StatusTypeRequested:
		req := msg.StatusType
		requestCounter.WithLabelValues(req).Inc()
		if _, exists := s.info.Requested[req]; exists {
			s.info.Requested[req]++
		} else {
			s.info.Requested[req] = 1
		}
	case StatusTypeAPIRequested:
		req := msg.StatusType
		apiCounter.WithLabelValues(req).Inc()
		if _, exists := s.info.APIRequested[req]; exists {
			s.info.APIRequested[req]++
		} else {
			s.info.APIRequested[req] = 1
		}
	case StatusTypeCacheUUID:
		req := msg.StatusType
		cacheCounter.WithLabelValues("cacheUUID", req).Inc()
		switch req {
		case "hit":
			s.info.CacheUUID.Hits++
		case "fresh":
			s.info.CacheUUID.FreshHits++
		case "stale":
			s.info.CacheUUID.StaleHits++
		case "miss":
			s.info.CacheUUID.Misses++
		case "error":
			s.info.CacheUUID.Errors++
		}
	case StatusTypeCacheUserData:
		req := msg.StatusType
		cacheCounter.WithLabelValues("cacheUserData", req).Inc()
		switch req {
		case "hit":
			s.info.CacheUserData.Hits++
		case "fresh":
			s.info.CacheUserData.FreshHits++
		case "stale":
			s.info.CacheUserData.StaleHits++
		case "miss":
			s.info.CacheUserData.Misses++
		case "error":
			s.info.CacheUserData.Errors++
		}
	case StatusTypeCacheSkin:
		req := msg.StatusType
		cacheCounter.WithLabelValues("cacheSkin", req).Inc()
		switch req {
		case "hit":
			s.info.CacheSkin.Hits++
		case "fresh":
			s.info.CacheSkin.FreshHits++
		case "stale":
			s.info.CacheSkin.StaleHits++
		case "miss":
			s.info.CacheSkin.Misses++
		case "error":
			s.info.CacheSkin.Errors++
		}
	case StatusTypeCacheSkinTransient:
		req := msg.StatusType
		cacheCounter.WithLabelValues("cacheSkinTransient", req).Inc()
		switch req {
		case "hit":
			s.info.CacheSkinTransient.Hits++
		case "fresh":
			s.info.CacheSkinTransient.FreshHits++
		case "stale":
			s.info.CacheSkinTransient.StaleHits++
		case "miss":
			s.info.CacheSkinTransient.Misses++
		case "error":
			s.info.CacheSkinTransient.Errors++
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

	s.info.ImgdMem = memstats.Alloc
	s.info.Uptime = time.Now().Unix() - s.StartedAt
	s.info.CacheUUID.Length = cache["cacheUUID"].Len()
	s.info.CacheUUID.Size = cache["cacheUUID"].Size()
	s.info.CacheUserData.Length = cache["cacheUserData"].Len()
	s.info.CacheUserData.Size = cache["cacheUserData"].Size()
	s.info.CacheSkin.Length = cache["cacheSkin"].Len()
	s.info.CacheSkin.Size = cache["cacheSkin"].Size()
	s.info.CacheSkinTransient.Length = cache["cacheSkinTransient"].Len()
	s.info.CacheSkinTransient.Size = cache["cacheSkinTransient"].Size()
}

// Increments the error counter for the specific type.
func (s *StatusCollector) Errored(errorType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeErrored,
		StatusType:  errorType,
	}
}

// Increments the request counter for the specific type.
func (s *StatusCollector) Requested(reqType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeRequested,
		StatusType:  reqType,
	}
}

// Increments the request counter for the specific type.
func (s *StatusCollector) APIRequested(reqType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeAPIRequested,
		StatusType:  reqType,
	}
}

// Increments the cache counter of the specifc type ("hit" or "miss").
func (s *StatusCollector) CacheUUID(statType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheUUID,
		StatusType:  statType,
	}
}

// Increments the cache counter of the specifc type ("hit" or "miss").
func (s *StatusCollector) CacheUserData(statType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheUserData,
		StatusType:  statType,
	}
}

// Increments the cache counter of the specifc type ("hit" or "miss").
func (s *StatusCollector) CacheSkin(statType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheSkin,
		StatusType:  statType,
	}
}

// Increments the cache counter of the specifc type ("hit" or "miss").
func (s *StatusCollector) CacheSkinTransient(statType string) {
	s.inputData <- statusCollectorMessage{
		MessageType: StatusTypeCacheSkinTransient,
		StatusType:  statType,
	}
}
