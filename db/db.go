package db

import (
	"fmt"

	dbtypes "github.com/samasno/lsm-storage-engine/types"
)

type DB struct {
	sequence uint64

	memtable        dbtypes.Memtable
	memtableInsertC chan Update
	memtableDoneC   chan struct{}

	manifest Manifest
}

type Session struct {
	Memtable dbtypes.Memtable
}

func NewDB(session Session) (*DB, error) {
	if session.Memtable == nil {
		return nil, fmt.Errorf("Session memtable cannot be nil")
	}

	db := &DB{
		memtable: session.Memtable,
	}

	return db, nil
}

func (db *DB) Start() {
	readyc := make(chan struct{})

	go db.runMemtable(readyc)
	<-readyc
}
