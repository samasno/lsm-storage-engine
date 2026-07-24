package lsmtree

import (
	"errors"
	"testing"

	"github.com/samasno/lsm-storage-engine/db"
)

func TestPut(t *testing.T) {
	lsm := NewLSM()

	err := lsm.Put([]byte("key"), []byte("value"))
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestPutNilKey(t *testing.T) {
	lsm := NewLSM()

	err := lsm.Put(nil, []byte("value"))
	if !errors.Is(err, ErrNilKey) {
		t.Fatalf("expected ErrNilKey, got %v", err)
	}
}

func TestGet(t *testing.T) {
	lsm := NewLSM()

	key := []byte("hello")
	value := []byte("world")

	err := lsm.Put(key, value)
	if err != nil {
		t.Fatal(err.Error())
	}

	got, err := lsm.Get(key)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(got) != string(value) {
		t.Fatalf("expected %s got %s", value, got)
	}
}

func TestGetNotFound(t *testing.T) {
	lsm := NewLSM()

	_, err := lsm.Get([]byte("missing"))
	if !errors.Is(err, db.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	lsm := NewLSM()

	key := []byte("hello")

	err := lsm.Put(key, []byte("world"))
	if err != nil {
		t.Fatal(err.Error())
	}

	err = lsm.Delete(key)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = lsm.Get(key)
	if !errors.Is(err, db.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestGetLatestVersion(t *testing.T) {
	lsm := NewLSM()

	key := []byte("hello")

	err := lsm.Put(key, []byte("first"))
	if err != nil {
		t.Fatal(err.Error())
	}

	err = lsm.Put(key, []byte("second"))
	if err != nil {
		t.Fatal(err.Error())
	}

	got, err := lsm.Get(key)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(got) != "second" {
		t.Fatalf("expected second got %s", got)
	}
}

func TestDeleteNilKey(t *testing.T) {
	lsm := NewLSM()

	err := lsm.Delete(nil)
	if !errors.Is(err, ErrNilKey) {
		t.Fatalf("expected ErrNilKey, got %v", err)
	}
}
