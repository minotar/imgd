package test_storage

import (
	"strconv"

	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/util/test_helpers"
)

type TestStoreBench struct {
	Keys  []string
	Store *TestStorage
	Size  int
}

func NewTestStoreBench() *TestStoreBench {
	return &TestStoreBench{
		Store: NewTestStorage(),
	}
}

func (tsb *TestStoreBench) MinSize(n int) {
	if tsb.Size > n {
		return
	}

	tsb.Keys = make([]string, n)
	for i := 0; i < n; i++ {
		tsb.Keys[i] = test_helpers.RandString(32)
		tsb.Store.Insert(tsb.Keys[i], []byte(strconv.Itoa(i)))
	}
}

// FillStore requires at least n mininum keys already in the TestStore - call MinSize first
func (tsb *TestStoreBench) FillStore(store storage.Storage, n int) {
	for i := 0; i < n; i++ {
		store.Insert(tsb.Keys[i], []byte(strconv.Itoa(i)))
	}

}

type TestStorage struct {
	store map[string][]byte
}

var _ storage.Storage = new(TestStorage)

func NewTestStorage() *TestStorage {
	return &TestStorage{
		store: make(map[string][]byte),
	}
}

func (m *TestStorage) Insert(key string, value []byte) error {
	m.store[string(key)] = value
	return nil
}

func (m *TestStorage) Retrieve(key string) ([]byte, error) {
	value, hit := m.store[string(key)]
	if !hit {
		return nil, storage.ErrNotFound
	}
	return value, nil
}

func (m *TestStorage) Flush() error {
	m.store = make(map[string][]byte)
	return nil
}

func (m *TestStorage) Len() uint {
	return uint(len(m.store))
}

// Size will not be accurate for an in-memory Store
func (m *TestStorage) Size() uint64 {
	return 0
}

func (m *TestStorage) Close() {
	return
}
