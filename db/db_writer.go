package db

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"
)

func (db *DB) Put(key, value []byte) error {
	return db.handleWrite(key, value, Insert)
}

func (db *DB) Delete(key []byte) error {
	return db.handleWrite(key, nil, Delete)
}

func (db *DB) CloseMemtable() {
	db.memtableDoneC <- struct{}{}
}

func (db *DB) handleWrite(key, value []byte, action Action) (err error) {
	assert(db.memtableInsertC != nil, "memtable insert channel not initialized, cannot add record")
	if nil == key {
		return fmt.Errorf("Key is nil")
	}

	update := Update{
		Key:       key,
		Value:     value,
		ResponseC: make(chan error),
		Action:    action,
		Sequence:  atomic.AddUint64(&db.sequence, 1),
	}

	db.memtableInsertC <- update

	err = <-update.ResponseC
	if err != nil {
		return err
	}

	return nil
}

// runs in own routine to mutliplex inserts
func (db *DB) runMemtable(readyc chan struct{}) {
	assert(db.memtable != nil, "starting insert loop with nil memtable")
	assert(db.memtableInsertC == nil, "memtable insert channel must be nil to ensure loop is not already running")
	assert(db.memtableDoneC == nil, "memtable done channel must be nil to ensure loop is not already running")

	db.memtableDoneC = make(chan struct{})
	db.memtableInsertC = make(chan Update, 1000)

	readyc <- struct{}{}

InsertLoop:
	for {
		select {
		case update := <-db.memtableInsertC:
			err := db.insertUpdate(update)
			update.ResponseC <- err
		case <-db.memtableDoneC:
			println("shutdown skiplist")
			close(db.memtableDoneC)
			close(db.memtableInsertC)
			db.memtableDoneC = nil
			db.memtableInsertC = nil
			break InsertLoop
		}
	}

}

func (db *DB) insertUpdate(update Update) error {
	encodedKey := encodeKey(update.Key, update.Action, update.Sequence)

	err := db.memtable.Insert(encodedKey, update.Value)
	if err != nil {
		return err
	}

	return nil
}

// keys must never be nil
func encodeKey(rawkey []byte, action Action, sequence uint64) []byte {
	assert(rawkey != nil && 0 != len(rawkey), "tried to format a nil or empty key")

	sequenceCode := sequence << 8
	sequenceCode |= uint64(action)
	rawKeyLen := len(rawkey)

	encodedKey := make([]byte, rawKeyLen+8)
	copy(encodedKey, rawkey)
	binary.LittleEndian.PutUint64(encodedKey[rawKeyLen:], sequenceCode)
	return encodedKey
}

func decodeKey(encodedKey []byte) (rawKey []byte, action Action, sequence uint64) {
	assert(encodedKey != nil || 9 < len(encodedKey), "received an invalid key to decode")
	rawKey = extractKeyRaw(encodedKey)
	sequence = extractKeySequenceNumber(encodedKey)
	action = extractKeyAction(encodedKey)

	return rawKey, action, sequence
}
