package memory

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// This is an attempt at a prefix tree . However, it's not used,
// since it seems to be about half the speed of using a simple
// map. It's also not (yet) very well tested.
type memoryTree struct {
	prefix   string
	children []*memoryTree
	value    []byte
}

// Returns the common prefix between this node's prefix, and
// the provided byte string.
func (m *memoryTree) commonPrefix(against string) int {
	l := min(len(against), len(m.prefix))
	i := 0
	for i < l && against[i] == m.prefix[i] {
		i++
	}

	return i
}

func (m *memoryTree) Delete(path string) {
	if len(path) == 0 {
		m.value = nil
		return
	}

	for _, child := range m.children {
		shared := child.commonPrefix(path)
		if shared == 0 {
			continue
		}

		child.Delete(path[shared:])
	}
}

func (m *memoryTree) Insert(path string, ptr []byte) {
	if len(path) == 0 {
		m.value = ptr
		return
	}

	for _, child := range m.children {
		shared := child.commonPrefix(path)
		if shared == 0 {
			continue
		}

		if shared == len(child.prefix) {
			child.value = ptr
		} else {
			child.Insert(path[shared:], ptr)
		}

		return
	}

	// No existing child found? We need to make a new one then.
	baby := newMemoryTree()
	baby.prefix = path
	baby.value = ptr
	m.children = append(m.children, baby)
}

func (m *memoryTree) Find(path string) []byte {
	if len(path) == 0 {
		return m.value
	}

	for _, child := range m.children {
		shared := child.commonPrefix(path)
		if shared == 0 {
			continue
		}

		return child.Find(path[shared:])
	}

	return nil
}

func newMemoryTree() *memoryTree {
	return &memoryTree{
		children: make([]*memoryTree, 0, 16),
		value:    nil,
	}
}
