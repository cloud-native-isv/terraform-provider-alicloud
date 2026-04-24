package alicloud

import "sync"

type MutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.Mutex),
	}
}

func (m *MutexKV) Lock(key string) {
	m.get(key).Lock()
}

func (m *MutexKV) Unlock(key string) {
	m.get(key).Unlock()
}

func (m *MutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()

	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}

	return mutex
}
