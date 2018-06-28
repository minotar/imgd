package bench

import (
	"github.com/minotar/imgd/storage"
	"github.com/minotar/imgd/storage/lru"
	"github.com/minotar/imgd/storage/memory"
	"github.com/minotar/imgd/storage/radix"
	"github.com/minotar/imgd/storage/redigo"
	"github.com/minotar/imgd/storage/util/helper"

	"testing"
	"time"
)

func benchInsBase(b *testing.B, c storage.Storage) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.Insert(helper.RandString(32), []byte(helper.RandString(32)), time.Hour); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	c.Flush()
}

func generateKeys(b *testing.B, c storage.Storage) []string {
	keys := []string{}
	for i := 0; i < b.N; i++ {
		key := helper.RandString(32)
		if err := c.Insert(key, []byte(helper.RandString(32)), time.Hour); err != nil {
			b.Fatal(err)
		}
		keys = append(keys, key)
	}

	return keys
}

func benchRetrieveBase(b *testing.B, c storage.Storage) {
	keys := generateKeys(b, c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Retrieve(keys[i]); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	c.Flush()
}

func BenchmarkMemoryInsert(b *testing.B) {
	benchInsBase(b, memory.New(b.N))
}

func BenchmarkLruInsert(b *testing.B) {
	benchInsBase(b, lru.New())
}

func BenchmarkRedigoInsert(b *testing.B) {
	benchInsBase(b, redigo.New("127.0.0.1:6379", ""))
}

func BenchmarkRadixInsert(b *testing.B) {
	store, _ := radix.New(radix.RedisConfig{
		Network: "tcp",
		Address: "127.0.0.1:6379",
		Size:    1,
	})
	benchInsBase(b, store)
}

func BenchmarkMemoryRetrieve(b *testing.B) {
	benchRetrieveBase(b, memory.New(b.N))
}

func BenchmarkLruRetrieve(b *testing.B) {
	benchRetrieveBase(b, lru.New())
}

func BenchmarkRedigoRetrieve(b *testing.B) {
	benchRetrieveBase(b, redigo.New("127.0.0.1:6379", ""))
}

func BenchmarkRadixRetrieve(b *testing.B) {
	store, _ := radix.New(radix.RedisConfig{
		Network: "tcp",
		Address: "127.0.0.1:6379",
		Size:    1,
	})
	benchRetrieveBase(b, store)
}
