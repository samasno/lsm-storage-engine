package dbtypes

import (
	"os"
)

type Memtable interface {
	Insert(key []byte, value []byte)
	Seek(seekkey []byte) (value []byte)
	SeekEqualOrLower(seekkey []byte) (key []byte, value []byte)
	Scanner(startKey []byte) Scanner
}

type Scanner interface {
	Next() bool
	Key() []byte
	Value() []byte
	Release() error
}

type Storage interface {
	Files() []*os.File
	NewFile(filename string) (StorageFile, error)
}

type StorageFile interface {
	Write(data []byte) (int, error)
	WriteAt(offset uint64, data []byte) error
	Read(buf []byte) (int, error)
	ReadAt(offset uint64, buf []byte) error
	Seek(whence int) (int, error)
}
