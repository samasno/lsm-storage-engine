package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	dbtypes "github.com/samasno/lsm-storage-engine/types"
)

const MagicNumber uint32 = 1828464582

type Manifest struct {
	storage dbtypes.Storage
	levels  []TreeLevel
}

type TreeLevel struct {
	Files []Datafile // must be sorted at all times by last index
}

func (m *Manifest) NewDatafile() (*Datafile, error) {
	// create new file at level 0
	filename := fmt.Sprintf("%d.sst", time.Now().UnixNano())
	sf, err := m.storage.NewFile(filename)
	if err != nil {
		return nil, err
	}

	df := &Datafile{
		Index: &DatablockIndex{},
		f:     sf,
	}

	return df, nil
}

type Datafile struct {
	f       dbtypes.StorageFile
	LastKey []byte
	Index   *DatablockIndex
}

func (df Datafile) Write(data []byte) (int, error) {
	return df.f.Write(data)
}

func GenerateFooter(indexLen, offset uint64) ([]byte, error) {
	output := bytes.NewBuffer([]byte{})
	err := binary.Write(output, binary.LittleEndian, indexLen)
	if err != nil {
		return nil, err
	}

	err = binary.Write(output, binary.LittleEndian, offset)
	if err != nil {
		return nil, err
	}

	err = binary.Write(output, binary.LittleEndian, MagicNumber)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

type DatablockIndex struct {
	index []DatablockIndexEntry
}

func (di *DatablockIndex) Append(entry DatablockIndexEntry) {
	di.index = append(di.index, entry)
}

func (di *DatablockIndex) Bytes() ([]byte, error) {
	output := bytes.NewBuffer([]byte{})
	for _, entry := range di.index {
		entryB, err := entry.Bytes()
		if err != nil {
			return nil, err
		}
		_, err = output.Write(entryB)
		if err != nil {
			return nil, err
		}
	}

	return output.Bytes(), nil
}

type DatablockIndexEntry struct {
	KeyLen  uint16
	LastKey []byte
	Offset  uint64
}

func NewDatablockIndexEntry(key []byte, offset uint64) DatablockIndexEntry {
	return DatablockIndexEntry{
		KeyLen:  uint16(len(key)),
		LastKey: key,
		Offset:  offset,
	}
}

func (di DatablockIndexEntry) Bytes() ([]byte, error) {
	output := bytes.NewBuffer([]byte{})
	err := binary.Write(output, binary.LittleEndian, di.KeyLen)
	if err != nil {
		return nil, err
	}

	_, err = output.Write(di.LastKey)
	if err != nil {
		return nil, err
	}

	err = binary.Write(output, binary.LittleEndian, di.Offset)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

type DatablockEntry struct {
	Key   []byte
	Value []byte
}

func NewDatablockEntry(key, value []byte) DatablockEntry {
	return DatablockEntry{
		Key:   key,
		Value: value,
	}
}

func (d DatablockEntry) Bytes() ([]byte, error) {
	if d.Key == nil {
		return nil, ErrNullKey
	}

	output := bytes.NewBuffer([]byte{})
	err := binary.Write(output, binary.LittleEndian, uint16(len(d.Key)))
	if err != nil {
		return nil, err
	}

	_, err = output.Write(d.Key)
	if err != nil {
		return nil, err
	}

	err = binary.Write(output, binary.LittleEndian, uint16(len(d.Value)))
	if err != nil {
		return nil, err
	}

	_, err = output.Write(d.Value)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

type Datablock struct {
	size      int
	max       int
	lastEntry DatablockEntry
	data      bytes.Buffer
}

func NewDataBlock() *Datablock {
	return &Datablock{
		max:  int(DatafileBlocksize),
		data: *bytes.NewBuffer([]byte{}),
	}
}

func (d *Datablock) Append(entry DatablockEntry) (bool, error) {
	dbytes, err := entry.Bytes()
	if err != nil {
		return false, err
	}

	dlen := len(dbytes)
	if d.max < d.size+dlen {
		return false, nil
	}

	_, err = d.data.Write(dbytes)
	if err != nil {
		return false, err
	}

	d.lastEntry = entry
	d.size += dlen

	return true, nil
}

func (d *Datablock) Bytes() []byte {
	if nil == d {
		return nil
	}

	return d.data.Bytes()
}

func (d *Datablock) LastKey() []byte {
	return d.lastEntry.Key
}
