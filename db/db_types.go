package db

type Action uint8

const (
	Insert Action = iota + 1
	Delete
	Seek
)

type Update struct {
	Key       []byte
	Value     []byte
	Action    Action
	Sequence  uint64
	ResponseC chan error
}

type Memtable interface {
	Insert(key []byte, value []byte)
	Seek(seekkey []byte) (value []byte)
	SeekEqualOrLower(seekkey []byte) (key []byte, value []byte)
}
