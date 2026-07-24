package db

import (
	"bytes"
	"errors"
	"fmt"
	"sync/atomic"
)

var ErrKeyNotFound = errors.New("key not found")

func (db *DB) Get(seekkey []byte) ([]byte, error) {
	if nil == seekkey || 0 == len(seekkey) {
		return nil, fmt.Errorf("Cannot GET nil or empty key")
	}

	sequence := atomic.LoadUint64(&db.sequence)
	encodedSeekKey := encodeKey(seekkey, Seek, sequence)

	// use seek equal or lower to catch all previous sequences with seekkey match
	encodedKey, value := db.memtable.SeekEqualOrLower(encodedSeekKey)

	if nil == encodedKey {
		return nil, ErrKeyNotFound
	}

	// if latest match for key is a delete, return nil
	rawKey, action, _ := decodeKey(encodedKey)
	if 0 != bytes.Compare(rawKey, seekkey) || Insert != action {
		return nil, ErrKeyNotFound
	}

	return value, nil
}
