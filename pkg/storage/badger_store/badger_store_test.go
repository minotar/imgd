package badger_store

import (
	"strconv"
	"testing"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_store"
)

const (
	TestBoltPath = "/tmp/badger"
)

func freshStore() *BadgerStore {
	store, _ := NewBadgerStore(TestBoltPath)
	store.Flush()
	return store
}

func TestRetrieveMiss(t *testing.T) {
	store := freshStore()
	defer store.Close()

	v, err := store.Retrieve(test_helpers.RandString(32))
	if v != nil {
		t.Errorf("Retrieve Miss should return a nil value, not: %+v", v)
	}
	if err != storage.ErrNotFound {
		t.Errorf("Retrieve Miss should return a storage.ErrNotFound, not: %s", err)
	}
}

func TestInsertAndRetrieve(t *testing.T) {
	store := freshStore()
	defer store.Close()

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.Insert(str, []byte(strconv.Itoa(i)))
		item, err := store.Retrieve(str)
		if err != nil {
			t.Errorf("Retrieve should not be an error: %s", err)
		}
		if string(item) != strconv.Itoa(i) {
			t.Errorf("%+v did not match %d", item, i)
		}
	}
}

func TestInsertAndDelete(t *testing.T) {
	store := freshStore()
	defer store.Close()

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
	store := freshStore()
	defer store.Close()

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.Insert(str, []byte("var"))
	}
}

var largeBucket = test_store.NewTestStoreBench()

func BenchmarkInsert(b *testing.B) {
	store := freshStore()
	defer store.Close()

	largeBucket.MinSize(b.N)
	b.ResetTimer()

	largeBucket.FillStore(store.Insert, b.N)
}

func BenchmarkLookup(b *testing.B) {
	store := freshStore()
	defer store.Close()

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
