package dbtypes

type Memtable interface {
	Insert(key []byte, value []byte)
	Seek(seekkey []byte) (value []byte)
	SeekEqualOrLower(seekkey []byte) (key []byte, value []byte)
}

type Scanner interface {
	Next() bool
	Key() []byte
	Value() []byte
	Release() error
}
