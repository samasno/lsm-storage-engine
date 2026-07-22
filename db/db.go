package db

import "fmt"

type DB struct {
	sequence uint64

	memtable        Memtable
	memtableInsertC chan Update
	memtableDoneC   chan struct{}
}

type Session struct {
	Memtable Memtable
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
