package store

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/cache/util/test_helpers"
)

type mockClock struct {
	time time.Time
}

func (m *mockClock) Now() time.Time {
	return m.time
}

func (m *mockClock) Add(t time.Duration) {
	m.time = m.time.Add(t)
}

func timeUTC() time.Time {
	mockedTime, _ := time.Parse(time.RFC3339, "2021-05-19T00:00:00Z")
	return mockedTime.UTC()
}

func unixUTC(n int) time.Time {
	return time.Unix(int64(n), 0).UTC()
}

func freshStoreExpiry() (*StoreExpiry, *mockClock) {
	se, _ := NewStoreExpiry(func() {}, time.Minute)
	clock := &mockClock{unixUTC(0)}
	se.Clock = clock
	return se, clock
}

func TestNewStoreNoCompactor(t *testing.T) {
	_, err := NewStoreExpiry(nil, 0)
	if err == nil || !strings.Contains(err.Error(), "Compactor function") {
		t.Errorf("Lack of specified Compactor function should have raised an error")
	}
}

func TestNewStoreExpiry(t *testing.T) {
	calledCount := 0
	compactorFunc := func() { calledCount++ }
	se, err := NewStoreExpiry(compactorFunc, 5*time.Millisecond)
	if err != nil {
		t.Error("NewStoreExpiry should not throw an error")
	}

	if calledCount != 0 {
		t.Errorf("compactorFunc shouldn't be called before Start()")
	}
	se.Start()
	time.Sleep(time.Duration(2) * time.Millisecond)
	if calledCount != 1 {
		t.Errorf("compactorFunc should be called once immediately after Start()")
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	se.Stop()
	time.Sleep(time.Duration(2) * time.Millisecond)

	// It should not be less than 2
	if calledCount < 2 {
		t.Errorf("compactorFunc should be called after ticking")
	}

}

func TestNewStoreEntry(t *testing.T) {
	se, clock := freshStoreExpiry()

	iterationCount := 500
	for i := 0; i < iterationCount; i++ {

		keyName := test_helpers.RandString(32)
		valueStr := test_helpers.RandString(256)
		e := se.NewStoreEntry(keyName, []byte(valueStr), time.Duration(i)*time.Minute)

		if keyName != e.Key {
			t.Errorf("Expected key \"%s\" did not match StoreEntry key \"%s\"", keyName, e.Key)
		}
		if bytes.Compare([]byte(valueStr), e.Value) == 1 {
			t.Error("Binary values did not match expected")
		}
		if i == 0 {
			if unixUTC(0) != e.Expiry() {
				t.Errorf("TTL Value of 0 should be Epoch 1970, not %s", e.Expiry())
			}
		} else if expectedTime := clock.Now().Add(time.Duration(i) * time.Minute); !expectedTime.Equal(e.Expiry()) {
			t.Errorf("Expected Time %s did not match StoreEntry Time %s", expectedTime, e.Expiry())
		}
	}

}

func TestStoreEntryExpiry(t *testing.T) {
	clock := &mockClock{unixUTC(1)}
	for i := 0; i < 256; i++ {
		buf := make([]byte, 4+i)

		// Incrementing by 1 second each time
		timeBytes := []byte{0, 0, 0, byte(i)}
		copy(buf, timeBytes)
		copy(buf[4:], []byte(test_helpers.RandString(i)))

		if len(buf) != 4+i {
			t.Errorf("Length of bytes should have been %d, not %d", 4+i, len(buf))
		}
		expirySeconds := getBytesExpirySeconds(timeBytes)
		if expirySeconds != uint32(i) {
			t.Errorf("Expected Expiry %d seconds, not %d", i, expirySeconds)
		}

		if i == 0 {
			if HasBytesExpired(timeBytes, clock.Now()) {
				t.Error("TTL Value of 0 should not be expiring")
			}
		} else {
			if HasBytesExpired(timeBytes, clock.Now()) {
				t.Errorf("Expiry %d *should not* be expired at %s", expirySeconds, clock.Now())
			}
			clock.Add(time.Duration(1) * time.Second)
			if !HasBytesExpired(timeBytes, clock.Now()) {
				t.Errorf("Expiry %d *should* be expired at %s", expirySeconds, clock.Now())
			}
		}
	}
}

func TestStoreEntryEncodeDecode(t *testing.T) {
	se, clock := freshStoreExpiry()

	iterationCount := 500
	for i := 0; i < iterationCount; i++ {

		keyName := test_helpers.RandString(32)
		valueStr := test_helpers.RandString(256)
		e := se.NewStoreEntry(keyName, []byte(valueStr), time.Duration(i)*time.Minute)

		keyBytes, valueBytes := e.Encode()

		e2 := DecodeStoreEntry(keyBytes, valueBytes)

		if keyName != e2.Key {
			t.Errorf("Expected key \"%s\" did not match BCE key \"%s\"", keyName, e.Key)
		}
		if !bytes.Equal([]byte(valueStr), e2.Value) {
			t.Error("Binary values did not match expected")
		}
		if i == 0 {
			if unixUTC(0) != e2.Expiry() {
				t.Errorf("TTL Value of 0 should be Epoch 1970, not %s", e.Expiry())
			}
		} else if expectedTime := clock.Now().Add(time.Duration(i) * time.Minute); !expectedTime.Equal(e.Expiry()) {
			t.Errorf("Expected Time %s did not match BCE Time %s", expectedTime, e.Expiry())
		}
	}
}

func benchEncode(size int, b *testing.B) {
	se, _ := freshStoreExpiry()

	keyStr := test_helpers.RandString(32)
	valueBytes := []byte(test_helpers.RandString(size))
	entry := se.NewStoreEntry(keyStr, valueBytes, time.Duration(b.N)*time.Microsecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1, r2 := entry.Encode()
		_, _ = r1, r2
	}
}

func benchDecode(size int, b *testing.B) {
	keyBytes := []byte(test_helpers.RandString(32))
	valueStr := test_helpers.RandString(size)

	buf := make([]byte, 4+size)
	copy(buf[:4], []byte{96, 164, 85, 0})
	copy(buf[4:], []byte(valueStr))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1 := DecodeStoreEntry(keyBytes, buf)
		_ = r1
	}
}

func benchEncodeDecode(size int, b *testing.B) {
	se, _ := freshStoreExpiry()

	keyStr := test_helpers.RandString(32)
	valueBytes := []byte(test_helpers.RandString(size))
	entry := se.NewStoreEntry(keyStr, valueBytes, time.Duration(b.N)*time.Microsecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyBytes, valueBytes := entry.Encode()
		r1 := DecodeStoreEntry(keyBytes, valueBytes)
		_ = r1
	}
}

func BenchmarkBoltCacheEntryEncode32(b *testing.B) {
	benchEncode(32, b)
}

func BenchmarkBoltCacheEntryEncode64(b *testing.B) {
	benchEncode(64, b)
}

func BenchmarkBoltCacheEntryEncode256(b *testing.B) {
	benchEncode(256, b)
}

func BenchmarkBoltCacheEntryEncode1024(b *testing.B) {
	benchEncode(1024, b)
}

func BenchmarkBoltCacheEntryDecode32(b *testing.B) {
	benchDecode(32, b)
}

func BenchmarkBoltCacheEntryDecode64(b *testing.B) {
	benchDecode(64, b)
}

func BenchmarkBoltCacheEntryDecode256(b *testing.B) {
	benchDecode(256, b)
}

func BenchmarkBoltCacheEntryDecode1024(b *testing.B) {
	benchDecode(1024, b)
}

func BenchmarkBoltCacheEntryEncodeDecode32(b *testing.B) {
	benchEncodeDecode(32, b)
}

func BenchmarkBoltCacheEntryEncodeDecode64(b *testing.B) {
	benchEncodeDecode(64, b)
}

func BenchmarkBoltCacheEntryEncodeDecode256(b *testing.B) {
	benchEncodeDecode(256, b)
}

func BenchmarkBoltCacheEntryEncodeDecode1024(b *testing.B) {
	benchEncodeDecode(1024, b)
}
