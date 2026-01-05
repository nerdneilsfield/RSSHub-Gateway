package cache

import (
	"container/list"
	"sync"
	"time"
)

type MemoryStore struct {
	enabled       bool
	maxItemBytes  int
	maxTotalBytes int

	mu         sync.Mutex
	entries    map[string]*list.Element
	order      *list.List
	totalBytes int
	strings    map[string]string
}

type memoryEntry struct {
	key       string
	value     Entry
	expiresAt time.Time
	size      int
}

func NewMemoryStore(opts Options) *MemoryStore {
	return &MemoryStore{
		enabled:       opts.Enabled,
		maxItemBytes:  opts.MaxItemBytes,
		maxTotalBytes: opts.MaxTotalBytes,
		entries:       make(map[string]*list.Element),
		order:         list.New(),
		strings:       make(map[string]string),
	}
}

func (m *MemoryStore) Provider() string {
	return "memory"
}

func (m *MemoryStore) GetResponse(key string) (Entry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	elem := m.entries[key]
	if elem == nil {
		return Entry{}, false
	}
	entry := elem.Value.(*memoryEntry)
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		m.removeElement(elem)
		return Entry{}, false
	}
	m.order.MoveToFront(elem)
	return cloneEntry(entry.value), true
}

func (m *MemoryStore) SetResponse(key string, entry Entry, ttl time.Duration) error {
	size := EntrySize(entry)
	if m.maxItemBytes > 0 && size > m.maxItemBytes {
		return ErrEntryTooLarge
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if elem := m.entries[key]; elem != nil {
		prev := elem.Value.(*memoryEntry)
		m.totalBytes -= prev.size
		prev.value = cloneEntry(entry)
		prev.expiresAt = expiresAt
		prev.size = size
		m.totalBytes += size
		m.order.MoveToFront(elem)
	} else {
		item := &memoryEntry{key: key, value: cloneEntry(entry), expiresAt: expiresAt, size: size}
		elem := m.order.PushFront(item)
		m.entries[key] = elem
		m.totalBytes += size
	}

	m.evict()
	return nil
}

func (m *MemoryStore) GetString(key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.strings[key]
	return value, ok, nil
}

func (m *MemoryStore) SetString(key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strings[key] = value
	return nil
}

func (m *MemoryStore) Close() error {
	return nil
}

func (m *MemoryStore) evict() {
	if m.maxTotalBytes <= 0 {
		return
	}
	for m.totalBytes > m.maxTotalBytes {
		back := m.order.Back()
		if back == nil {
			return
		}
		m.removeElement(back)
	}
}

func (m *MemoryStore) removeElement(elem *list.Element) {
	entry := elem.Value.(*memoryEntry)
	delete(m.entries, entry.key)
	m.order.Remove(elem)
	m.totalBytes -= entry.size
}

func cloneEntry(entry Entry) Entry {
	out := Entry{Status: entry.Status}
	if len(entry.Headers) > 0 {
		out.Headers = make([]Header, len(entry.Headers))
		copy(out.Headers, entry.Headers)
	}
	if len(entry.Body) > 0 {
		out.Body = append([]byte(nil), entry.Body...)
	}
	return out
}
