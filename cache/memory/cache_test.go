package memory

import (
	"math/rand"
	"strconv"
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

func TestInsertAndFind(t *testing.T) {
	tree := New()
	for i := 0; i < 10; i++ {
		str := randString(32)
		tree.Insert(str, []byte(strconv.Itoa(i)), time.Second)
		item, _ := tree.Find(str)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

type testBucket struct {
	keys  []string
	cache *MemoryCache
	size  int
}

var largeBucket testBucket

func initLargeBucket(n int) {
	if largeBucket.size > n {
		return
	}
	keys := make([]string, n)
	cache := New()
	for i := 0; i < n; i++ {
		keys[i] = randString(32)
		cache.Insert(keys[i], []byte(strconv.Itoa(i)), time.Second)
	}

	largeBucket = testBucket{keys, cache, n}
}

func BenchmarkInsert(b *testing.B) {
	initLargeBucket(b.N)
	tree := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(largeBucket.keys[i], []byte(strconv.Itoa(i)), time.Second)
	}
}

func BenchmarkLookup(b *testing.B) {
	initLargeBucket(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < 10; k++ {
			largeBucket.cache.Find(largeBucket.keys[k])
		}
	}
}
