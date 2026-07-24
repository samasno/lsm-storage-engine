package lsmtree

import (
	"errors"

	"github.com/samasno/lsm-storage-engine/db"
	skiplist "github.com/samasno/lsm-storage-engine/memtable"
)

type LogStructuredMergeTree struct {
	db *db.DB
}

var ErrNilKey = errors.New("Nil or empty key provided")

func NewLSM() *LogStructuredMergeTree {
	memtable := skiplist.NewSkipList(skiplist.SortKeysDescending)

	sess := db.Session{
		Memtable: memtable,
	}

	lsm := &LogStructuredMergeTree{}

	var err error
	lsm.db, err = db.NewDB(sess)
	if err != nil {
		panic(err.Error())
	}

	lsm.db.Start()

	return lsm
}

func (lsm *LogStructuredMergeTree) Put(key, value []byte) error {
	if nilOrEmpty(key) {
		return ErrNilKey
	}

	err := lsm.db.Put(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (lsm *LogStructuredMergeTree) Delete(key []byte) error {
	if nilOrEmpty(key) {
		return ErrNilKey
	}

	err := lsm.db.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func (lsm *LogStructuredMergeTree) Get(key []byte) ([]byte, error) {
	if nilOrEmpty(key) {
		return nil, ErrNilKey
	}

	value, err := lsm.db.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func nilOrEmpty(key []byte) bool {
	if nil == key || 0 == len(key) {
		return true
	}

	return false
}
