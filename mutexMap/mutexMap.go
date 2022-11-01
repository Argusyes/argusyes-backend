package mutexMap

import "sync"

type MutexMap[T any] struct {
	mp map[string]T
	sync.RWMutex
}

func NewMutexMap[T any](size int) MutexMap[T] {
	return MutexMap[T]{
		mp: make(map[string]T, size),
	}
}

func (m *MutexMap[T]) Get(key string) (T, bool) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.mp[key]
	return v, ok
}

func (m *MutexMap[T]) GetNilThenSet(key string, v T) T {
	m.Lock()
	defer m.Unlock()
	r, ok := m.mp[key]
	if ok {
		return r
	} else {
		m.mp[key] = v
		return v
	}

}

func (m *MutexMap[T]) Set(key string, v T) {
	m.Lock()
	defer m.Unlock()
	m.mp[key] = v
}

func (m *MutexMap[T]) Remove(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.mp, key)
}

func (m *MutexMap[T]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.mp)
}

func (m *MutexMap[T]) Has(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.mp[key]
	return ok
}

func (m *MutexMap[T]) Each(f func(key string, v T)) {
	m.RLock()
	defer m.RUnlock()
	for key, val := range m.mp {
		f(key, val)
	}
}
