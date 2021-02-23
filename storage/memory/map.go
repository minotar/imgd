package memory

import "github.com/minotar/imgd/storage"

// Simple memory cache store that uses a plain Go map.
type memoryMap struct {
	m    map[string][]byte
	size int
}

func (m *memoryMap) Retrieve(path string) ([]byte, error) {
	e, hit := m.m[string(path)]
	if !hit {
		return nil, storage.ErrNotFound
	}
	return e, nil
}

// Todo: Fix this.
func (m *memoryMap) Insert(path string, ptr []byte) {
	if len(m.m) >= m.size {
		// Delete a "random" skin when we are full
		for k := range m.m {
			delete(m.m, k)
			break
		}
	}

	m.m[string(path)] = ptr
}

func (m *memoryMap) Delete(path string) {
	delete(m.m, path)
}

func (m *memoryMap) Len() uint {
	return uint(len(m.m))
}

func newMemoryMap(maxEntries int) *memoryMap {
	return &memoryMap{
		m:    make(map[string][]byte, maxEntries),
		size: maxEntries,
	}
}
