// Common methods for Expiry packages
package expiry

import (
	"fmt"
	"time"
)

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (r realClock) Now() time.Time { return time.Now() }

// Handles tracking of expiration times.
// This is intentionally _not_ an interface
// The way the Expiry packages implement this is unique
type Expiry struct {
	Clock             clock
	compactorFunc     func()
	closer            chan bool
	compactorInterval time.Duration
	running           bool
}

type Options struct {
	Clock             clock
	CompactorFunc     func()
	CompactorInterval time.Duration
}

var DefaultOptions = &Options{
	Clock:             realClock{},
	CompactorInterval: time.Duration(5) * time.Minute,
}

func NewExpiry(options *Options) (*Expiry, error) {
	var e = &Expiry{closer: make(chan bool)}

	e.Clock = options.Clock
	e.compactorFunc = options.CompactorFunc
	e.compactorInterval = options.CompactorInterval

	if e.compactorFunc == nil {
		// If either function is missing, then throw an error
		return nil, fmt.Errorf("missing Expiry Compactor function")
	}
	return e, nil
}

func (e *Expiry) Start() {
	if !e.running {
		go e.runCompactor()
		e.running = true
	}
}

func (e *Expiry) Stop() {
	if e.running {
		// Signal to the runCompactor it should stop ticking
		e.closer <- true
		// Setting to false so an in-progress compaction can stop
		e.running = false
	}
}

// runCompactor is in its own goroutine and thus needs the closer to stop
func (e *Expiry) runCompactor() {
	// Run immediately
	e.compactorFunc()
	ticker := time.NewTicker(e.compactorInterval)

COMPACT:
	for {
		select {
		case <-e.closer:
			break COMPACT
		case <-ticker.C:
			e.compactorFunc()
		}
	}

	ticker.Stop()
}

func (e *Expiry) NewExpiryRecordTTL(key string, ttl time.Duration) ExpiryRecord {
	return NewExpiryRecordTTL(key, e.Clock, ttl)
}
