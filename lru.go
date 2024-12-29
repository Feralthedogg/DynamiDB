package main

import (
	"sync"
)

type LRUCacheItem struct {
	key   string
	value []byte
	prev  *LRUCacheItem
	next  *LRUCacheItem
}

type LRUCache struct {
	capacity int
	items    map[string]*LRUCacheItem

	head *LRUCacheItem
	tail *LRUCacheItem

	mu sync.RWMutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*LRUCacheItem),
	}
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	node, found := c.items[key]
	c.mu.RUnlock()
	if !found {
		return nil, false
	}

	c.mu.Lock()
	c.moveToFront(node)
	value := node.value
	c.mu.Unlock()

	return value, true
}

func (c *LRUCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, found := c.items[key]; found {
		node.value = value
		c.moveToFront(node)
	} else {
		newItem := &LRUCacheItem{
			key:   key,
			value: value,
		}
		c.items[key] = newItem
		c.addToFront(newItem)

		if len(c.items) > c.capacity {
			c.removeOldest()
		}
	}
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, found := c.items[key]; found {
		c.removeNode(node)
		delete(c.items, key)
	}
}

func (c *LRUCache) moveToFront(node *LRUCacheItem) {
	if node == c.head {
		return
	}
	c.removeNode(node)
	c.addToFront(node)
}

func (c *LRUCache) addToFront(node *LRUCacheItem) {
	node.prev = nil
	node.next = c.head
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRUCache) removeNode(node *LRUCacheItem) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}

func (c *LRUCache) removeOldest() {
	if c.tail == nil {
		return
	}
	oldest := c.tail
	c.removeNode(oldest)
	delete(c.items, oldest.key)
}
