package store

import (
	"bytes"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
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

func TestNewBoltCacheEntry(t *testing.T) {
	clock := &mockClock{timeUTC()}
	bc := &BoltCache{
		clock: clock,
	}

	iterationCount := 500
	for i := 0; i < iterationCount; i++ {
		keyStr := "key_%s" + test_helpers.RandString(32)
		valueStr := test_helpers.RandString(256)

		bce := bc.NewBoltCacheEntry(keyStr, []byte(valueStr), time.Duration(i)*time.Minute)

		if keyStr != bce.Key {
			t.Errorf("Expected key \"%s\" did not match BCE key \"%s\"", keyStr, bce.Key)
		}
		if bytes.Compare([]byte(valueStr), bce.Value) == 1 {
			t.Error("Binary values did not match expected")
		}
		if i == 0 {
			if unixUTC(0) != bce.Expiry() {
				t.Errorf("TTL Value of 0 should be Epoch 1970, not %s", bce.Expiry())
			}
		} else if expectedTime := clock.Now().Add(time.Duration(i) * time.Minute); !expectedTime.Equal(bce.Expiry()) {
			t.Errorf("Expected Time %s did not match BCE Time %s", expectedTime, bce.Expiry())
		}
	}

}

func TestBoltCacheEntryGetExpiry(t *testing.T) {
	for i := 0; i < 256; i++ {
		buf := make([]byte, i+4)

		// Incrementing by 1 second each time
		timeBytes := []byte{0, 0, 0, byte(i)}
		copy(buf, timeBytes)
		copy(buf[4:], []byte(test_helpers.RandString(0)))

		if len(buf) != 4+i {
			t.Errorf("Length of bytes should have been %d, not %d", i+4, len(buf))
		}
		expiryTime := getExpiry(getExpirySeconds(timeBytes))
		if expectedTime := time.Unix(int64(i), 0); !expectedTime.Equal(expiryTime) {
			t.Errorf("Expected Time %+v did not Expiry Time %+v", expectedTime, expiryTime)
		}
	}
}

func TestBoltCacheEntryEncodeDecode(t *testing.T) {
	clock := &mockClock{timeUTC()}
	bc := &BoltCache{
		clock: clock,
	}

	iterationCount := 500
	for i := 0; i < iterationCount; i++ {
		keyStr := "key_%s" + test_helpers.RandString(32)
		valueStr := test_helpers.RandString(256)
		bce := bc.NewBoltCacheEntry(keyStr, []byte(valueStr), time.Duration(i)*time.Minute)

		keyBytes, valueBytes := bce.Encode()

		bce2 := DecodeBoltCacheEntry(keyBytes, valueBytes)

		if keyStr != bce2.Key {
			t.Errorf("Expected key \"%s\" did not match BCE key \"%s\"", keyStr, bce.Key)
		}
		if bytes.Compare([]byte(valueStr), bce2.Value) == 1 {
			t.Error("Binary values did not match expected")
		}
		if i == 0 {
			if unixUTC(0) != bce2.Expiry() {
				t.Errorf("TTL Value of 0 should be Epoch 1970, not %s", bce.Expiry())
			}
		} else if expectedTime := clock.Now().Add(time.Duration(i) * time.Minute); !expectedTime.Equal(bce.Expiry()) {
			t.Errorf("Expected Time %s did not match BCE Time %s", expectedTime, bce.Expiry())
		}
		//fmt.Printf("Expected Time %s did not match BCE Time %s\n", expectedTime, bce.Expiry())
	}
}

func benchEncode(size int, b *testing.B) {
	bc := &BoltCache{
		clock: &mockClock{unixUTC(0)},
	}
	keyStr := test_helpers.RandString(32)
	valueStr := test_helpers.RandString(size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bce := bc.NewBoltCacheEntry(keyStr, []byte(valueStr), time.Duration(b.N)*time.Microsecond)
		_, _ = bce.Encode()
	}
}

func benchDecode(size int, b *testing.B) {
	key := []byte(test_helpers.RandString(32))
	valueStr := test_helpers.RandString(size)

	buf := make([]byte, 4+size)
	copy(buf[:4], []byte{96, 164, 85, 0})
	copy(buf[4:], []byte(valueStr))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodeBoltCacheEntry(key, buf)
	}
}

func benchEncodeDecode(size int, b *testing.B) {
	bc := &BoltCache{
		clock: &mockClock{unixUTC(0)},
	}
	keyStr := test_helpers.RandString(32)
	valueStr := test_helpers.RandString(size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bce := bc.NewBoltCacheEntry(keyStr, []byte(valueStr), time.Duration(b.N)*time.Microsecond)
		key, value := bce.Encode()
		_ = DecodeBoltCacheEntry(key, value)
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
