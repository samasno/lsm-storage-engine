package skiplist

import (
	"bytes"
	"math/rand"
	"sync"
)

const maxHeight uint8 = 32

type Skiplist struct {
	head       [maxHeight]*SkipListNode
	mtx        *sync.RWMutex
	maxHeight  uint8
	height     uint8
	comparator Comparator
}

type Comparator func(a, b []byte) int

type SkipListNode struct {
	next   [maxHeight]*SkipListNode
	key    []byte
	value  []byte
	height uint8
}

func NewSkipList(comparator Comparator) *Skiplist {
	return &Skiplist{
		maxHeight:  maxHeight,
		mtx:        &sync.RWMutex{},
		head:       [maxHeight]*SkipListNode{},
		comparator: comparator,
	}
}

func newNode(key, value []byte) *SkipListNode {
	return &SkipListNode{key: key, value: value, next: [maxHeight]*SkipListNode{}}
}

func (sk *Skiplist) Insert(key, value []byte) (err error) {
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

		comp := sk.comparator(head.key, node.key)

		if 1 == comp {
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

			compNext := sk.comparator(current.next[level].key, node.key)
			if 1 == compNext {
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

func (sk *Skiplist) SeekEqualOrLower(seekkey []byte) (key []byte, value []byte) {
	assert(seekkey != nil && 0 != len(seekkey), "Cannot seek nil or empty key")
	if nil == sk.head[0] {
		return nil, nil
	}

	level := sk.height
	var current *SkipListNode
	for {
		if 0 == level {
			break
		}

		current = sk.head[level]
		comp := sk.comparator(current.key, seekkey)
		if 0 == comp {
			return current.key, current.value
		}

		if 1 == comp {
			level--
			continue
		}

		break
	}

	// traverse list, descend until match or hit level 0
	for {
		if 0 == level {
			break
		}

		comp := sk.comparator(current.key, seekkey)
		if 0 == comp {
			return current.key, current.value
		}

		if nil == current.next[level] {
			level--
			continue
		}

		compNext := sk.comparator(current.next[level].key, seekkey)
		if -1 == compNext {
			current = current.next[level]
			continue
		} else {
			level--
			continue
		}
	}

	// traverse list until lte is found
	for {
		if nil == current {
			break
		}

		comp := sk.comparator(current.key, seekkey)
		if -1 == comp {
			current = current.next[0]
			continue
		}

		return current.key, current.value
	}

	return nil, nil
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

// returns 1 if incoming is less than existing
func SortKeysAscending(existing, incoming []byte) int {
	return bytes.Compare(existing, incoming)
}

// returns 1 if incoming is greater than existing
func SortKeysDescending(existing, incoming []byte) int {
	return bytes.Compare(incoming, existing)
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
