package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"
)

func TestDBPutsRecords(t *testing.T) {
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
		db.Put([]byte(key), []byte(key+key))
	}

	mock, ok := mm.(*mockMemtable)
	if !ok {
		t.Fatal("failed to cast memtable")
	}

	results := make([]bool, 100)

	for k := range mock.data {
		_, action, sequence := decodeKey([]byte(k))

		if action != Insert {
			t.Fatalf("expected action %d got %d", Insert, action)
		}

		results[sequence-1] = true
	}

	// ensures that every sequence inserted is incremented
	for i, ok := range results {
		if !ok {
			t.Fatalf("did not find entry for %d", i)
		}
	}
}

func TestDBInsertsDeletes(t *testing.T) {
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
		db.Delete([]byte(key))
	}

	mock, ok := mm.(*mockMemtable)
	if !ok {
		t.Fatal("failed to cast memtable")
	}

	results := make([]bool, 100)

	for k := range mock.data {
		_, action, sequence := decodeKey([]byte(k))

		assertEqual(t, "Should decode original action", Action(action), Delete)

		results[sequence-1] = true
	}

	for i, ok := range results {
		if !ok {
			t.Fatalf("did not find entry for %d", i)
		}
	}
}

func TestEncodeKey(t *testing.T) {
	key := "testing"
	action := Insert
	sequence := uint64(1)
	encoded := encodeKey([]byte(key), action, sequence)

	resKey := encoded[:len(encoded)-8]
	if string(resKey) != key {
		t.Fatalf("expected %s got %s", key, string(resKey))
	}

	sequenceCodeB := encoded[len(encoded)-8:]
	var sequenceCode uint64
	binary.Read(bytes.NewBuffer(sequenceCodeB), binary.LittleEndian, &sequenceCode)

	decodedSequence := sequenceCode >> 8

	assertEqual(t, "Decodes to original sequence", decodedSequence, sequence)

	decodedAction := Action(sequenceCode)

	assertEqual(t, "Decodes to original action", decodedAction, action)
}

func TestDecodeKey(t *testing.T) {
	key := "testing"
	action := Insert
	sequence := uint64(1)
	encoded := encodeKey([]byte(key), action, sequence)

	dkey, daction, dsequence := decodeKey(encoded)

	assertEqual(t, "Decodes to original key", string(dkey), key)
	assertEqual(t, "Decodes to original action", daction, action)
	assertEqual(t, "Decodes to origin sequence", dsequence, sequence)

}

type mockMemtable struct {
	data map[string][]byte
	mtx  *sync.Mutex
}

func newMockMemtable() Memtable {
	return &mockMemtable{
		data: map[string][]byte{},
		mtx:  &sync.Mutex{},
	}
}

func (mm *mockMemtable) Insert(key []byte, value []byte) error {
	mm.mtx.Lock()
	defer mm.mtx.Unlock()
	mm.data[string(key)] = value
	return nil
}

func assertEqual[T comparable](t *testing.T, name string, actual, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: got %v, expected %v", name, actual, expected)
	}
}
