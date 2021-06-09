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
	if calledCount != 1 {
		t.Errorf("compactorFunc should be called once immediately after Start()")
	}
	time.Sleep(time.Duration(4) * time.Millisecond)
	expiry.Stop()

	if calledCount != 2 {
		t.Errorf("compactorFunc should be called after ticking")
	}

}
