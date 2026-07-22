package lsmtree

import (
	"fmt"

	"github.com/samasno/lsm-storage-engine/db"
	skiplist "github.com/samasno/lsm-storage-engine/memtable"
)

type LogStructuredMergeTree struct {
	db *db.DB
}

func NewLSM() *LogStructuredMergeTree {
	memtable := skiplist.NewSkipList()

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
	if nil == key || 0 == len(key) {
		return fmt.Errorf("Cannot PUT nil or empty key")
	}

	err := lsm.db.Put(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (lsm *LogStructuredMergeTree) Delete(key []byte) error {
	if nil == key || 0 == len(key) {
		return fmt.Errorf("Cannot DELETE nil or empty key")
	}

	err := lsm.db.Delete(key)
	if err != nil {
		return err
	}

	return nil
}
