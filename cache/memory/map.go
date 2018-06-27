package memory

// Simple memory cache store that uses a plain Go map.
type memoryMap struct {
	m map[string][]byte
}

func (m *memoryMap) Find(path string) []byte {
	e, _ := m.m[string(path)]
	return e
}

func (m *memoryMap) Insert(path string, ptr []byte) {
	m.m[string(path)] = ptr
}

func (m *memoryMap) Delete(path string) {
	delete(m.m, path)
}

func newMemoryMap() *memoryMap {
	return &memoryMap{m: make(map[string][]byte)}
}
