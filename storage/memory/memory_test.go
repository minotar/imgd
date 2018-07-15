package memory

import (
	"strconv"
	"testing"
	"time"

	"github.com/minotar/imgd/storage/util/helper"
)

func TestInsertAndFind(t *testing.T) {
	cache, _ := New(10)
	for i := 0; i < 10; i++ {
		str := helper.RandString(32)
		cache.Insert(str, []byte(strconv.Itoa(i)), time.Minute)
		item, _ := cache.Retrieve(str)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

func TestSize(t *testing.T) {
	cache, _ := New(10)
	// Iterate 20 times with a cache size of just 10
	for i := 0; i < 20; i++ {
		str := helper.RandString(32)
		cache.Insert(str, []byte(strconv.Itoa(i)), time.Minute)
	}
	if cache.Len() != 10 {
		t.Fail()
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
	cache, _ := New(n)
	for i := 0; i < n; i++ {
		keys[i] = helper.RandString(32)
		cache.Insert(keys[i], []byte(strconv.Itoa(i)), time.Minute)
	}

	largeBucket = testBucket{keys, cache, n}
}

func BenchmarkInsert(b *testing.B) {
	initLargeBucket(b.N)
	cache, _ := New(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Insert(largeBucket.keys[i], []byte(strconv.Itoa(i)), time.Minute)
	}
}

func BenchmarkLookup(b *testing.B) {
	initLargeBucket(b.N)

	iter := 10
	if b.N < 10 {
		iter = b.N
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < iter; k++ {
			largeBucket.cache.Retrieve(largeBucket.keys[k])
		}
	}
}
