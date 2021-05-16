package lru

import (
	"strconv"
	"testing"

	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_storage"
)

func TestInsertAndRetrieve(t *testing.T) {
	store, _ := NewLruStore(10)
	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.Insert(str, []byte(strconv.Itoa(i)))
		item, _ := store.Retrieve(str)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

func TestHousekeeping(t *testing.T) {
	store, _ := NewLruStore(5)

	if len := store.Len(); len != 0 {
		t.Errorf("Initialized store should be length 0, not %d", len)
	}

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.Insert(str, []byte("var"))
	}

	if len := store.Len(); len != 5 {
		t.Errorf("Full store should be length 5, not %d", len)
	}

	store.Flush()

	if len := store.Len(); len != 0 {
		t.Errorf("Flushed store should be length 0, not %d", len)
	}
}

var largeBucket = test_storage.NewTestStoreBench()

func BenchmarkInsert(b *testing.B) {
	store, _ := NewLruStore(b.N)

	largeBucket.MinSize(b.N)
	b.ResetTimer()

	largeBucket.FillStore(store, b.N)
}

func BenchmarkLookup(b *testing.B) {
	store, _ := NewLruStore(b.N)

	// Set TestBucket and Store based on a static size (b.N should only affect loop)
	largeBucket.MinSize(1000)
	largeBucket.FillStore(store, 1000)

	// Each operation we will read the same set of keys
	iter := 10
	if b.N < 10 {
		iter = b.N
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for k := 0; k < iter; k++ {
			store.Retrieve(largeBucket.Keys[k])
		}
	}
}
