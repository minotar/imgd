package cache

import (
	"github.com/minotar/imgd/cache/memory"
	"github.com/minotar/imgd/cache/redis"

	"math/rand"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func benchInsBase(b *testing.B, c Cache) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.Insert(randString(32), []byte(randString(32)), time.Hour); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	c.Flush()
}

func generateKeys(b *testing.B, c Cache) []string {
	keys := []string{}
	for i := 0; i < b.N; i++ {
		key := randString(32)
		if err := c.Insert(key, []byte(randString(32)), time.Hour); err != nil {
			b.Fatal(err)
		}
		keys = append(keys, key)
	}

	return keys
}

func benchFindBase(b *testing.B, c Cache) {
	keys := generateKeys(b, c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Find(keys[i]); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	c.Flush()
}

func BenchmarkMemoryInsert(b *testing.B) {
	benchInsBase(b, memory.New())
}

func BenchmarkRedisInsert(b *testing.B) {
	benchInsBase(b, redis.New("127.0.0.1:6379", ""))
}

func BenchmarkMemoryFind(b *testing.B) {
	benchFindBase(b, memory.New())
}

func BenchmarkRedisFind(b *testing.B) {
	benchFindBase(b, redis.New("127.0.0.1:6379", ""))
}
