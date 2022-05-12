package expiry

import (
	"strings"
	"testing"
	"time"
)

func TestExpiryNoCompactor(t *testing.T) {
	_, err := NewExpiry(DefaultOptions)
	if err == nil || !strings.Contains(err.Error(), "Compactor function") {
		t.Errorf("Lack of specified Compactor function should have raised an error")
	}
}

func TestExpiryRealClock(t *testing.T) {
	expiryOptions := DefaultOptions
	expiryOptions.CompactorFunc = func() {}
	expiry, _ := NewExpiry(DefaultOptions)

	r := expiry.NewExpiryRecordTTL("foo", time.Minute)

	if r.HasExpired(time.Now()) {
		t.Error("Created record should not have expired")
	}
}

func TestCompactor(t *testing.T) {
	calledCount := 0

	expiryOptions := DefaultOptions
	expiryOptions.CompactorFunc = func() { calledCount++ }
	expiryOptions.CompactorInterval = 5 * time.Millisecond
	expiry, _ := NewExpiry(DefaultOptions)

	if calledCount != 0 {
		t.Errorf("compactorFunc shouldn't be called before Start()")
	}
	expiry.Start()
	time.Sleep(time.Duration(2) * time.Millisecond)

	// It should be exactly 1
	if calledCount != 1 {
		t.Errorf("compactorFunc should be called once immediately after Start()")
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	expiry.Stop()
	time.Sleep(time.Duration(2) * time.Millisecond)

	// It should not be less than 2
	if calledCount < 2 {
		t.Errorf("compactorFunc should be called after ticking")
	}

}
