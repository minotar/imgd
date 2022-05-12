package legacy_storage

import (
	"testing"
	"time"

	"github.com/minotar/imgd/storage/util/helper"
)

/*
func TestInsertAndRetrieve(t *testing.T) {
	cache := &DebugCache{cache: make(map[string][]byte)}
	steveSkin, _ := minecraft.FetchSkinForSteve()
	InsertSkin(cache, "steve", steveSkin)
	fromCache, _ := RetrieveSkin(cache, "steve")
	if steveSkin.Hash != fromCache.Hash {
		t.Fail()
	}
}

func TestRetrieveMiss(t *testing.T) {
	cache := &DebugCache{cache: make(map[string][]byte)}
	fromCache, err := RetrieveSkin(cache, helper.RandString(32))
	if fromCache.Hash == err.Error() {
		t.Fail()
	}
}
*/

func TestInsertAndRetrieveKV(t *testing.T) {
	cache := &DebugCache{cache: make(map[string][]byte)}
	for i := 0; i < 10; i++ {
		key := helper.RandString(32)
		value := helper.RandString(32)
		InsertKV(cache, key, value, time.Minute)
		stringValue, _ := RetrieveKV(cache, key)
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
	cache := &DebugCache{cache: make(map[string][]byte)}
	for i := 0; i < 10; i++ {
		key := helper.RandString(32)
		startStruct := &testStruct{
			Name:      helper.RandString(32),
			ID:        i,
			Timestamp: time.Now(),
		}
		InsertGob(cache, key, startStruct, time.Minute)
		endStruct := &testStruct{}
		RetrieveGob(cache, key, endStruct)
		if startStruct.Name != endStruct.Name {
			t.Fail()
		}
		if startStruct.ID != endStruct.ID {
			t.Fail()
		}
		if !startStruct.Timestamp.Equal(endStruct.Timestamp) {
			t.Logf("\nS: %s\nE: %s", startStruct.Timestamp.Format(time.RFC3339Nano), startStruct.Timestamp.Format(time.RFC3339Nano))
			t.Fail()
		}
	}
}

func TestRetrieveGobMiss(t *testing.T) {
	cache := &DebugCache{cache: make(map[string][]byte)}
	key := helper.RandString(32)
	name := helper.RandString(32)
	startStruct := &testStruct{
		Name: name,
	}
	RetrieveGob(cache, key, startStruct)
	if startStruct.Name != name {
		t.Fail()
	}
}

type DebugCache struct {
	cache map[string][]byte
}

func (m *DebugCache) Insert(key string, value []byte, ttl time.Duration) error {
	m.cache[string(key)] = value
	return nil
}

func (m *DebugCache) Retrieve(key string) ([]byte, error) {
	value, hit := m.cache[string(key)]
	if !hit {
		return nil, ErrNotFound
	}
	return value, nil

}

func (m *DebugCache) Flush() error {
	m.cache = make(map[string][]byte)
	return nil
}

func (m *DebugCache) Len() uint {
	return uint(len(m.cache))
}

// Size will not be accurate for an in-memory Cache
func (m *DebugCache) Size() uint64 {
	return 0
}

func (m *DebugCache) Close() {
	return
}
