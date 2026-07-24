package db

import (
	"errors"
	"fmt"
	"testing"
)

func TestGetKey(t *testing.T) {
	mm := newMockMemtable()

	sess := Session{
		Memtable: mm,
	}

	db, err := NewDB(sess)
	if err != nil {
		t.Fatal(err.Error())
	}

	db.Start()

	for i := range 100 {
		key := fmt.Sprintf("%d", i)
		db.Put([]byte(key), []byte(key))
	}

	testValue := "22"

	value, err := db.Get([]byte("22"))
	if err != nil {
		t.Fatal(err.Error())
	}

	assertEqual(t, "DB gets correct value", string(value), testValue)
}

func TestGetKeyNotFound(t *testing.T) {
	mm := newMockMemtable()

	sess := Session{Memtable: mm}

	db, err := NewDB(sess)
	if err != nil {
		t.Fatal(err.Error())
	}

	db.Start()

	_, err = db.Get([]byte("missing"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestGetAfterDelete(t *testing.T) {
	mm := newMockMemtable()

	sess := Session{Memtable: mm}

	db, err := NewDB(sess)
	if err != nil {
		t.Fatal(err.Error())
	}

	db.Start()

	key := []byte("hello")

	err = db.Put(key, []byte("world"))
	if err != nil {
		t.Fatal(err.Error())
	}

	err = db.Delete(key)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = db.Get(key)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestGetNilKey(t *testing.T) {
	mm := newMockMemtable()

	sess := Session{Memtable: mm}

	db, err := NewDB(sess)
	if err != nil {
		t.Fatal(err.Error())
	}

	db.Start()

	_, err = db.Get(nil)
	if err == nil {
		t.Fatal("expected error for nil key, got nil")
	}
}
