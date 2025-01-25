package main

import (
	"sync"
	"time"
)

type TTLManager struct {
	cache     *LRUCache
	skiplist  *SkipList
	nodeByKey map[string]*skipListNode

	mu sync.RWMutex
}

func NewTTLManager(cache *LRUCache) *TTLManager {
	m := &TTLManager{
		cache:     cache,
		skiplist:  NewSkipList(),
		nodeByKey: make(map[string]*skipListNode),
	}
	go m.runCleaner()
	return m
}

func (m *TTLManager) SetExpire(key string, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldNode, ok := m.nodeByKey[key]; ok {
		m.skiplist.Remove(oldNode.expireAt, oldNode.key)
	}

	newNode, _ := m.skiplist.Insert(t, key)
	m.nodeByKey[key] = newNode
}

func (m *TTLManager) DeleteExpire(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node, ok := m.nodeByKey[key]; ok {
		m.skiplist.Remove(node.expireAt, node.key)
		delete(m.nodeByKey, key)
	}
}

func (m *TTLManager) IsExpired(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodeByKey[key]
	if !ok {
		return false
	}
	return time.Now().After(node.expireAt)
}

func (m *TTLManager) runCleaner() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		m.mu.Lock()
		for {
			earliest := m.skiplist.GetEarliest()
			if earliest == nil {
				break
			}
			if earliest.expireAt.After(now) {

				break
			}
			m.cache.Delete(earliest.key)

			removed := m.skiplist.RemoveEarliest()
			if removed != nil {
				delete(m.nodeByKey, removed.key)
			}
		}
		m.mu.Unlock()
	}
}
