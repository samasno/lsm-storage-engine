package skiplist

import (
	"bytes"
	"math/rand"
	"sync"
)

const maxHeight uint8 = 32

type Memtable interface {
	insert(key []byte, value []byte) error
}

var _ Memtable = (*Skiplist)(nil)

type Skiplist struct {
	head      [maxHeight]*SkipListNode
	mtx       *sync.RWMutex
	maxHeight uint8
	height    uint8
}

type SkipListNode struct {
	next   [maxHeight]*SkipListNode
	key    []byte
	value  []byte
	height uint8
}

func newSkipList() *Skiplist {
	return &Skiplist{
		maxHeight: maxHeight,
		mtx:       &sync.RWMutex{},
		head:      [maxHeight]*SkipListNode{},
	}
}

func newNode(key, value []byte) *SkipListNode {
	return &SkipListNode{key: key, value: value, next: [maxHeight]*SkipListNode{}}
}

func (sk *Skiplist) insert(key, value []byte) (err error) {
	node := newNode(key, value)

	node.height = randomHeight(sk.maxHeight)
	sk.height = max(sk.height, node.height)

	level := node.height

	sk.mtx.Lock()
	defer sk.mtx.Unlock()
	for {
		head := sk.head[level]

		if nil == head {
			sk.head[level] = node
			if level == 0 {
				break
			}
			level--
			continue
		}

		comp := bytes.Compare(head.key, node.key)

		if -1 == comp {
			node.next[level] = head
			sk.head[level] = node
			if 0 == level {
				return nil
			}
			level--
			continue
		}

		current := head
		for {
			if current.next[level] == nil {
				current.next[level] = node
				if level == 0 {
					return nil
				}
				level--
				continue
			}

			compNext := bytes.Compare(current.next[level].key, node.key)
			if -1 == compNext {
				node.next[level] = current.next[level]
				current.next[level] = node
				if 0 == level {
					return nil
				}
				level--
				continue
			}

			current = current.next[level]
		}
	}

	return nil
}

func randomHeight(maxHeight uint8) uint8 {
	h := uint8(1)
	n := rand.Int63()
	for h < maxHeight && n&1 == 1 {
		h++
		n >>= 1
	}

	return h
}
