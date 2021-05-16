package storage_test

import (
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_storage"
)

func TestInsertAndRetrieveKV(t *testing.T) {
	cache := test_storage.NewTestStorage()
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		value := test_helpers.RandString(32)
		storage.InsertKV(cache, key, value)
		stringValue, _ := storage.RetrieveKV(cache, key)
		if stringValue != value {
			t.Fail()
		}
		byteValue, _ := cache.Retrieve(key)
		if string(byteValue) != value {
			t.Fail()
		}
	}
}

type testStruct struct {
	Name      string
	ID        int
	Timestamp time.Time
}

func TestInsertAndRetrieveGob(t *testing.T) {
	cache := test_storage.NewTestStorage()
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		startStruct := &testStruct{
			Name:      test_helpers.RandString(32),
			ID:        i,
			Timestamp: time.Now(),
		}
		storage.InsertGob(cache, key, startStruct)
		endStruct := &testStruct{}
		storage.RetrieveGob(cache, key, endStruct)
		if startStruct.Name != endStruct.Name {
			t.Fail()
		}
		if startStruct.ID != endStruct.ID {
			t.Fail()
		}
		if !startStruct.Timestamp.Equal(endStruct.Timestamp) {
			t.Errorf("\nS: %s\nE: %s", startStruct.Timestamp.Format(time.RFC3339Nano), startStruct.Timestamp.Format(time.RFC3339Nano))
		}
	}
}

func TestRetrieveGobMiss(t *testing.T) {
	cache := test_storage.NewTestStorage()
	key := test_helpers.RandString(32)
	name := test_helpers.RandString(32)
	startStruct := &testStruct{
		Name: name,
	}
	storage.RetrieveGob(cache, key, startStruct)
	if startStruct.Name != name {
		t.Fail()
	}
}
