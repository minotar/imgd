package bolt_store

import (
	"strconv"
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_store"
)

const (
	TestBoltPath       = "/tmp/bolt_store_test.db"
	TestBoltBucketName = "bolt_test"
)

func freshStore() *BoltStore {
	store, _ := NewBoltStore(TestBoltPath, TestBoltBucketName)
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

func TestBatchInsertAndRetrieve(t *testing.T) {
	store := freshStore()
	defer store.Close()

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.InsertBatch(str, []byte(strconv.Itoa(i)))
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

	if len := store.Len(); len != 0 {
		t.Errorf("Initialized store should be length 0, not %d", len)
	}

	for i := 0; i < 10; i++ {
		str := test_helpers.RandString(32)
		store.Insert(str, []byte("var"))
	}

	if len := store.Len(); len != 10 {
		t.Errorf("Part filled store should be length 10, not %d", len)
	}

	store.Flush()

	if len := store.Len(); len != 0 {
		t.Errorf("Flushed store should be length 0, not %d", len)
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

func BenchmarkInsertNoSync(b *testing.B) {
	store := freshStore()
	defer store.Close()

	largeBucket.MinSize(b.N)
	store.DB.NoSync = true
	b.ResetTimer()

	largeBucket.FillStore(store.Insert, b.N)
}

func BenchmarkInsertParallel(b *testing.B) {
	store := freshStore()
	defer store.Close()

	largeBucket.MinSize(b.N)

	insertQueue := make(chan int, b.N)
	for count := 0; count < b.N; count++ {
		insertQueue <- count
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := <-insertQueue
			store.Insert(largeBucket.Keys[i], []byte(strconv.Itoa(i)))
		}
	})
}

func BenchmarkBatchInsertParallel(b *testing.B) {
	store := freshStore()
	defer store.Close()

	largeBucket.MinSize(b.N)

	store.DB.MaxBatchDelay = 20 * time.Millisecond
	insertQueue := make(chan int, b.N)
	for count := 0; count < b.N; count++ {
		insertQueue <- count
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := <-insertQueue
			store.InsertBatch(largeBucket.Keys[i], []byte(strconv.Itoa(i)))
		}
	})
}

func BenchmarkBatchInsertParallelNoSync(b *testing.B) {
	store := freshStore()
	defer store.Close()

	largeBucket.MinSize(b.N)

	store.DB.NoSync = true
	store.DB.MaxBatchSize = 3000
	store.DB.MaxBatchDelay = 5 * time.Millisecond
	insertQueue := make(chan int, b.N)
	for count := 0; count < b.N; count++ {
		insertQueue <- count
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := <-insertQueue
			store.InsertBatch(largeBucket.Keys[i], []byte(strconv.Itoa(i)))
		}
	})
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
