package skiplist

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func TestInsertsInOrder(t *testing.T) {
	sk := newSkipList()

	count := 100
	generateTestEntries(count, sk)

	err := crossLevelConsistency(sk)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = skiplistLevelsOrdered(sk)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func generateTestEntries(count int, sk *Skiplist) error {
	for i := 0; i < count; i++ {
		key := make([]byte, 20)
		_, err := rand.Read(key)
		if err != nil {
			return err
		}

		value := make([]byte, 20)
		_, err = rand.Read(value)
		if err != nil {
			return err
		}
		sk.Insert(key, value)
		// println("inserted", i)
	}

	return nil
}

func assert(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Error(message)
	}
}

func assertEqual[T comparable](t *testing.T, name string, actual, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: got %v, expected %v", name, actual, expected)
	}
}

func skiplistLevelsOrdered(sk *Skiplist) error {
	for i := uint8(0); i <= sk.height; i++ {
		err := checkLevelOrdered(sk, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkLevelOrdered(sk *Skiplist, level uint8) error {
	current := sk.head[level]

	if nil == current {
		return fmt.Errorf("head of skiplist is nil")
	}

	i := 0
	for {
		if nil == current.next[level] {
			break
		}

		if -1 == bytes.Compare(current.key, current.next[level].key) {
			return fmt.Errorf("Found out of order element at level %d index %d", level, i)
		}

		current = current.next[level]
		i++
	}

	return nil
}

func crossLevelConsistency(sk *Skiplist) error {
	for upper := uint8(1); upper <= sk.height; upper++ {
		err := levelIsSubsequence(sk, upper, upper-1)
		if err != nil {
			return err
		}
	}

	return nil
}

func levelIsSubsequence(sk *Skiplist, upper, lower uint8) error {
	lo := sk.head[lower]

	for hi := sk.head[upper]; hi != nil; hi = hi.next[upper] {
		for lo != nil && lo != hi {
			lo = lo.next[lower]
		}
		if lo == nil {
			return fmt.Errorf("node %v at level %d missing from level %d", hi.key, upper, lower)
		}
	}

	return nil
}
