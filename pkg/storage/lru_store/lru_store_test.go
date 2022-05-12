package lru_store

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	test_store "github.com/minotar/imgd/pkg/storage/util/test_store"
)

func TestRetrieveMiss(t *testing.T) {
	store, _ := NewLruStore(10)

	v, err := store.Retrieve(test_helpers.RandString(32))
	if v != nil {
		t.Errorf("Retrieve Miss should return a nil value, not: %+v", v)
	}
	if err != storage.ErrNotFound {
		t.Errorf("Retrieve Miss should return a storage.ErrNotFound, not: %s", err)
	}
}

func TestInsertAndRetrieve(t *testing.T) {
	store, _ := NewLruStore(10)
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		store.Insert(key, []byte(strconv.Itoa(i)))
		item, _ := store.Retrieve(key)
		if string(item) != strconv.Itoa(i) {
			t.Fail()
		}
	}
}

func TestInsertAndDelete(t *testing.T) {
	store, _ := NewLruStore(10)
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		store.Insert(key, []byte(strconv.Itoa(i)))
		store.Remove(key)
		_, err := store.Retrieve(key)
		if err != storage.ErrNotFound {
			t.Errorf("Key should have been removed: %s", key)
		}
	}
}

func TestHousekeeping(t *testing.T) {
	store, _ := NewLruStore(5)

	if len := store.Len(); len != 0 {
		t.Errorf("Initialized store should be length 0, not %d", len)
	}

	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		store.Insert(key, []byte("var"))
	}

	if len := store.Len(); len != 5 {
		t.Errorf("Full store should be length 5, not %d", len)
	}

	store.Flush()

	if len := store.Len(); len != 0 {
		t.Errorf("Flushed store should be length 0, not %d", len)
	}
}

func TestLruEviction(t *testing.T) {
	store, _ := NewLruStore(5)

	var keys [10]string
	for i := 0; i < 10; i++ {
		keys[i] = test_helpers.RandString(32)
	}

	// Fill 6 keys - should evict first key
	for i := 0; i < 6; i++ {
		store.Insert(keys[i], []byte("var"))
	}

	// Verify first key was evicted
	if _, err := store.Retrieve(keys[0]); err != storage.ErrNotFound {
		t.Errorf("First added key should have been evicted %s", keys[0])
	}

	// Verify second key is still present - and bump it's listing (as recently used)
	if _, err := store.Retrieve(keys[1]); err == storage.ErrNotFound {
		t.Errorf("Second key should not have been evicted %s", keys[0])
	}

	// Fill 3 more keys - should evict third/fourth/fifth keys - not second
	for i := 6; i < 9; i++ {
		fmt.Printf("Key id is %d\n", i)
		store.Insert(keys[i], []byte("var"))

		// Verify specific keys were evicted
		if _, err := store.Retrieve(keys[i-4]); err != storage.ErrNotFound {
			t.Errorf("keys[%d] should have been evicted when adding keys[%d]", i-4, i)
		}
	}

	// Verify second key is still present
	if _, err := store.Retrieve(keys[1]); err == storage.ErrNotFound {
		t.Errorf("Second key should not have been evicted %s", keys[0])
	}
}

var largeBucket = test_store.NewTestStoreBench()

func BenchmarkInsert(b *testing.B) {
	store, _ := NewLruStore(b.N)

	largeBucket.MinSize(b.N)
	b.ResetTimer()

	largeBucket.FillStore(store.Insert, b.N)
}

func BenchmarkLookup(b *testing.B) {
	store, _ := NewLruStore(b.N)

	// Set TestBucket and Store based on a static size (b.N should only affect loop)
	largeBucket.MinSize(1000)
	largeBucket.FillStore(store.Insert, 1000)

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
