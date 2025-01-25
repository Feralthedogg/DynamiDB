// slab.go

package main

import (
	"sync"
)

type Slab struct {
	chunkSize  int
	freeChunks chan []byte
}

type MultiSlabManager struct {
	slabs []Slab
	mu    sync.Mutex
}

var slabSizes = []int{64, 128, 256, 1024, 4096}

func NewMultiSlabManager() *MultiSlabManager {
	manager := &MultiSlabManager{}
	for _, size := range slabSizes {
		s := Slab{
			chunkSize:  size,
			freeChunks: make(chan []byte, 1000),
		}
		manager.slabs = append(manager.slabs, s)
	}
	return manager
}

func (m *MultiSlabManager) Allocate(size int) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.slabs {
		if m.slabs[i].chunkSize >= size {
			select {
			case chunk := <-m.slabs[i].freeChunks:
				return chunk[:size]
			default:
				return make([]byte, size, m.slabs[i].chunkSize)
			}
		}
	}
	return make([]byte, size)
}

func (m *MultiSlabManager) Free(buf []byte) {
	capacity := cap(buf)
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.slabs {
		if m.slabs[i].chunkSize == capacity {
			select {
			case m.slabs[i].freeChunks <- buf[:capacity]:
			default:
			}
			return
		}
	}
}

func (m *MultiSlabManager) Defragment() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var totalFree int
	for i := range m.slabs {
		totalFree += len(m.slabs[i].freeChunks)
	}

	if totalFree > 3000 {
		for i := range m.slabs {
			freeCount := len(m.slabs[i].freeChunks)
			toDiscard := freeCount / 2
			for j := 0; j < toDiscard; j++ {
				select {
				case <-m.slabs[i].freeChunks:
				default:
				}
			}
		}
	}
}
