package bigcache

import (
	"strconv"
	"testing"
	"time"

	"github.com/minotar/imgd/storage/util/helper"
)

func TestInsertAndRetrieve(t *testing.T) {
	cache, _ := New()
	for i := 0; i < 10; i++ {
		str := helper.RandString(32)
		cache.Insert(str, []byte(strconv.Itoa(i)), time.Second)
		item, _ := cache.Retrieve(str)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

type testBucket struct {
	keys  []string
	cache *BigcacheCache
	size  int
}

var largeBucket testBucket

func initLargeBucket(n int) {
	if largeBucket.size > n {
		return
	}
	keys := make([]string, n)
	cache, _ := New()
	for i := 0; i < n; i++ {
		keys[i] = helper.RandString(32)
		cache.Insert(keys[i], []byte(strconv.Itoa(i)), time.Second)
	}

	largeBucket = testBucket{keys, cache, n}
}

func BenchmarkInsert(b *testing.B) {
	initLargeBucket(b.N)
	cache, _ := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Insert(largeBucket.keys[i], []byte(strconv.Itoa(i)), time.Second)
	}
}

func BenchmarkLookup(b *testing.B) {
	initLargeBucket(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < 10; k++ {
			largeBucket.cache.Retrieve(largeBucket.keys[k])
		}
	}
}
