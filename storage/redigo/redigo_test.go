package redigo

import (
	"strconv"
	"testing"
	"time"

	"github.com/minotar/imgd/storage/util/helper"
)

func TestInsertAndFind(t *testing.T) {
	cache := New("127.0.0.1:6379", "")
	preSize := cache.Size()
	for i := 0; i < 10; i++ {
		str := helper.RandString(32)
		cache.Insert(str, []byte(strconv.Itoa(i)), time.Minute)
		item, _ := cache.Retrieve(str)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
	if cache.Len() != 10 {
		t.Fail()
	}
	if cache.Size() < preSize {
		t.Fail()
	}
}

type testBucket struct {
	keys  []string
	cache *RedisCache
	size  int
}

var largeBucket testBucket

func initLargeBucket(n int) {
	if largeBucket.size > n {
		return
	}
	keys := make([]string, n)
	cache := New("127.0.0.1:6379", "")
	for i := 0; i < n; i++ {
		keys[i] = helper.RandString(32)
		cache.Insert(keys[i], []byte(strconv.Itoa(i)), time.Minute)
	}

	largeBucket = testBucket{keys, cache, n}
}

func BenchmarkInsert(b *testing.B) {
	initLargeBucket(b.N)
	cache := New("127.0.0.1:6379", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Insert(largeBucket.keys[i], []byte(strconv.Itoa(i)), time.Minute)
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
