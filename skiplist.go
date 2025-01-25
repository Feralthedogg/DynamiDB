package main

import (
	"math/rand"
	"time"
)

const (
	maxLevel = 16
	pFactor  = 0.25
)

type skipListNode struct {
	expireAt time.Time
	key      string
	forward  []*skipListNode
}

type SkipList struct {
	head  *skipListNode
	level int
	size  int
}

func NewSkipList() *SkipList {
	return &SkipList{
		head: &skipListNode{
			forward: make([]*skipListNode, maxLevel),
		},
		level: 1,
		size:  0,
	}
}

func randomLevel() int {
	level := 1
	for rand.Float64() < pFactor && level < maxLevel {
		level++
	}
	return level
}

func less(timeA time.Time, keyA string, timeB time.Time, keyB string) bool {
	if timeA.Before(timeB) {
		return true
	} else if timeA.After(timeB) {
		return false
	}
	return keyA < keyB
}

func equal(timeA time.Time, keyA string, timeB time.Time, keyB string) bool {
	return timeA.Equal(timeB) && keyA == keyB
}

func (sl *SkipList) Insert(expireAt time.Time, key string) (newNode *skipListNode, replaced bool) {
	update := make([]*skipListNode, maxLevel)
	cur := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for cur.forward[i] != nil &&
			less(cur.forward[i].expireAt, cur.forward[i].key, expireAt, key) {
			cur = cur.forward[i]
		}
		update[i] = cur
	}

	next := cur.forward[0]
	if next != nil && next.key == key {
		sl.Remove(next.expireAt, next.key)
		replaced = true
	}

	newLvl := randomLevel()
	if newLvl > sl.level {
		for i := sl.level; i < newLvl; i++ {
			update[i] = sl.head
		}
		sl.level = newLvl
	}

	newNode = &skipListNode{
		expireAt: expireAt,
		key:      key,
		forward:  make([]*skipListNode, newLvl),
	}
	for i := 0; i < newLvl; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
	sl.size++
	return newNode, replaced
}

func (sl *SkipList) Remove(expireAt time.Time, key string) bool {
	update := make([]*skipListNode, maxLevel)
	cur := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for cur.forward[i] != nil &&
			less(cur.forward[i].expireAt, cur.forward[i].key, expireAt, key) {
			cur = cur.forward[i]
		}
		update[i] = cur
	}

	target := cur.forward[0]
	if target != nil && equal(target.expireAt, target.key, expireAt, key) {
		for i := 0; i < sl.level; i++ {
			if update[i].forward[i] == target {
				update[i].forward[i] = target.forward[i]
			}
		}
		for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
			sl.level--
		}
		sl.size--
		return true
	}
	return false
}

func (sl *SkipList) GetEarliest() *skipListNode {
	return sl.head.forward[0]
}

func (sl *SkipList) RemoveEarliest() *skipListNode {
	first := sl.head.forward[0]
	if first == nil {
		return nil
	}
	for i := 0; i < sl.level; i++ {
		if sl.head.forward[i] == first {
			sl.head.forward[i] = first.forward[i]
		}
	}
	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}
	sl.size--
	return first
}

func (sl *SkipList) Len() int {
	return sl.size
}
