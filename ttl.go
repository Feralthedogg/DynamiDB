package main

import (
	"sync"
	"time"
)

type TTLManager struct {
	cache    *LRUCache
	expireAt map[string]time.Time
	mu       sync.RWMutex
}

func NewTTLManager(cache *LRUCache) *TTLManager {
	m := &TTLManager{
		cache:    cache,
		expireAt: make(map[string]time.Time),
	}
	go m.runCleaner()
	return m
}

func (m *TTLManager) SetExpire(key string, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expireAt[key] = t
}

func (m *TTLManager) DeleteExpire(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.expireAt, key)
}

func (m *TTLManager) IsExpired(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if t, ok := m.expireAt[key]; ok {
		return time.Now().After(t)
	}
	return false
}

func (m *TTLManager) runCleaner() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for k, t := range m.expireAt {
			if now.After(t) {
				m.cache.Delete(k)
				delete(m.expireAt, k)
			}
		}
		m.mu.Unlock()
	}
}
