package storage_test

import (
	"testing"
	"time"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
	"github.com/minotar/imgd/pkg/storage/util/test_store"
)

func TestInsertAndRetrieveKV(t *testing.T) {
	store := test_store.NewTestStorage()
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		value := test_helpers.RandString(32)
		storage.InsertKV(store, key, value)
		stringValue, _ := storage.RetrieveKV(store, key)
		if stringValue != value {
			t.Fail()
		}
		byteValue, _ := store.Retrieve(key)
		if string(byteValue) != value {
			t.Fail()
		}
	}
}

func TestInsertAndDeleteKV(t *testing.T) {
	store := test_store.NewTestStorage()
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		storage.InsertKV(store, key, "foobar")
		store.Remove(key)
		_, err := store.Retrieve(key)
		if err != storage.ErrNotFound {
			t.Errorf("Key should have been removed: %s", key)
		}
	}
}

type testStruct struct {
	Name      string
	ID        int
	Timestamp time.Time
}

func TestInsertAndRetrieveGob(t *testing.T) {
	store := test_store.NewTestStorage()
	for i := 0; i < 10; i++ {
		key := test_helpers.RandString(32)
		startStruct := &testStruct{
			Name:      test_helpers.RandString(32),
			ID:        i,
			Timestamp: time.Now(),
		}
		storage.InsertGob(store, key, startStruct)
		endStruct := &testStruct{}
		storage.RetrieveGob(store, key, endStruct)
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
	store := test_store.NewTestStorage()
	key := test_helpers.RandString(32)
	name := test_helpers.RandString(32)
	startStruct := &testStruct{
		Name: name,
	}
	storage.RetrieveGob(store, key, startStruct)
	if startStruct.Name != name {
		t.Fail()
	}
}
