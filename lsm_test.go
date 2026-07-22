package lsmtree

import "testing"

func TestInsert(t *testing.T) {
	lsm := NewLSM()

	err := lsm.Put([]byte("test"), []byte("works"))
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDelete(t *testing.T) {
	lsm := NewLSM()

	err := lsm.Delete([]byte("test"))
	if err != nil {
		t.Fatal(err.Error())
	}
}
