package db

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"testing"

	dbtypes "github.com/samasno/lsm-storage-engine/types"
)

type mockStorageFile struct {
	buf bytes.Buffer
}

func (m *mockStorageFile) Write(data []byte) (int, error)        { return m.buf.Write(data) }
func (m *mockStorageFile) WriteAt(offset uint64, data []byte) error { return nil }
func (m *mockStorageFile) Read(buf []byte) (int, error)          { return m.buf.Read(buf) }
func (m *mockStorageFile) ReadAt(offset uint64, buf []byte) error { return nil }
func (m *mockStorageFile) Seek(whence int) (int, error)          { return 0, nil }

type mockStorage struct {
	file *mockStorageFile
}

func newMockStorage() *mockStorage {
	return &mockStorage{file: &mockStorageFile{}}
}

func (ms *mockStorage) Files() []*os.File { return nil }

func (ms *mockStorage) NewFile(_ string) (dbtypes.StorageFile, error) {
	return ms.file, nil
}

type flushKV struct {
	key, value []byte
}

type mockFlushScanner struct {
	entries []flushKV
	pos     int
}

func (s *mockFlushScanner) Next() bool {
	s.pos++
	return s.pos <= len(s.entries)
}

func (s *mockFlushScanner) Key() []byte    { return s.entries[s.pos-1].key }
func (s *mockFlushScanner) Value() []byte  { return s.entries[s.pos-1].value }
func (s *mockFlushScanner) Release() error { return nil }

type mockFlushMemtable struct {
	entries []flushKV
}

func (m *mockFlushMemtable) Insert(key, value []byte)                     {}
func (m *mockFlushMemtable) Seek(key []byte) []byte                       { return nil }
func (m *mockFlushMemtable) SeekEqualOrLower(key []byte) ([]byte, []byte) { return nil, nil }
func (m *mockFlushMemtable) Scanner(_ []byte) dbtypes.Scanner {
	return &mockFlushScanner{entries: m.entries}
}

func newFlushDB(entries []flushKV) (*DB, *mockStorageFile) {
	ms := newMockStorage()
	db := &DB{
		memtable: &mockFlushMemtable{entries: entries},
		manifest: Manifest{storage: ms},
	}
	return db, ms.file
}

func buildEntry(key, value []byte) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, uint16(len(key)))
	buf.Write(key)
	binary.Write(buf, binary.LittleEndian, uint16(len(value)))
	buf.Write(value)
	return buf.Bytes()
}

func buildIndexEntry(key []byte, offset uint64) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, uint16(len(key)))
	buf.Write(key)
	binary.Write(buf, binary.LittleEndian, offset)
	return buf.Bytes()
}

func buildFooter(indexLen, offset uint64) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, indexLen)
	binary.Write(buf, binary.LittleEndian, offset)
	binary.Write(buf, binary.LittleEndian, MagicNumber)
	return buf.Bytes()
}

// TestFlushSingleBlock verifies entry encoding and index/footer structure
// when all entries fit within one data block.
func TestFlushSingleBlock(t *testing.T) {
	entries := []flushKV{
		{[]byte("aaa"), []byte("val1")},
		{[]byte("bbb"), []byte("val2")},
		{[]byte("ccc"), []byte("val3")},
	}

	db, sf := newFlushDB(entries)

	if err := db.FlushMemtableToDisk(); err != nil {
		t.Fatal(err)
	}

	written := sf.buf.Bytes()

	var dataBlock []byte
	for _, e := range entries {
		dataBlock = append(dataBlock, buildEntry(e.key, e.value)...)
	}

	index := buildIndexEntry([]byte("ccc"), 0)
	footer := buildFooter(uint64(len(index)), DatafileBlocksize)

	expected := append(dataBlock, index...)
	expected = append(expected, footer...)

	if !bytes.Equal(written, expected) {
		t.Fatalf("written bytes mismatch: got %d bytes, want %d bytes", len(written), len(expected))
	}
}

// TestFlushMultipleBlocks verifies that entries exceeding a block's capacity
// are split across blocks with correct index entries and offsets.
func TestFlushMultipleBlocks(t *testing.T) {
	// entry size: 2+3+2+1000 = 1007 bytes
	// 4 entries = 4028 bytes ≤ 4096; 5th (total 5035) exceeds limit → two blocks
	largeVal := []byte(strings.Repeat("x", 1000))
	entries := []flushKV{
		{[]byte("k01"), largeVal},
		{[]byte("k02"), largeVal},
		{[]byte("k03"), largeVal},
		{[]byte("k04"), largeVal},
		{[]byte("k05"), largeVal},
	}

	db, sf := newFlushDB(entries)

	if err := db.FlushMemtableToDisk(); err != nil {
		t.Fatal(err)
	}

	written := sf.buf.Bytes()

	var block1 []byte
	for _, e := range entries[:4] {
		block1 = append(block1, buildEntry(e.key, e.value)...)
	}
	block2 := buildEntry(entries[4].key, entries[4].value)

	idx1 := buildIndexEntry([]byte("k04"), 0)
	idx2 := buildIndexEntry([]byte("k05"), DatafileBlocksize)
	index := append(idx1, idx2...)

	footer := buildFooter(uint64(len(index)), 2*DatafileBlocksize)

	var expected []byte
	expected = append(expected, block1...)
	expected = append(expected, block2...)
	expected = append(expected, index...)
	expected = append(expected, footer...)

	if !bytes.Equal(written, expected) {
		t.Fatalf("written bytes mismatch: got %d bytes, want %d bytes", len(written), len(expected))
	}
}

// TestFlushFooterMagicNumber parses the footer from raw bytes and verifies
// the magic number, index length, and offset fields.
func TestFlushFooterMagicNumber(t *testing.T) {
	entries := []flushKV{
		{[]byte("key"), []byte("value")},
	}

	db, sf := newFlushDB(entries)

	if err := db.FlushMemtableToDisk(); err != nil {
		t.Fatal(err)
	}

	written := sf.buf.Bytes()

	// footer layout: indexLen(8) + offset(8) + magic(4) = 20 bytes
	if len(written) < 20 {
		t.Fatalf("written data too short: %d bytes", len(written))
	}

	footerB := written[len(written)-20:]
	indexLen := binary.LittleEndian.Uint64(footerB[:8])
	offset := binary.LittleEndian.Uint64(footerB[8:16])
	magic := binary.LittleEndian.Uint32(footerB[16:])

	assertEqual(t, "magic number", magic, MagicNumber)
	assertEqual(t, "footer offset", offset, DatafileBlocksize)

	// one index entry: keyLen(2) + "key"(3) + offset(8) = 13 bytes
	assertEqual(t, "index length", indexLen, uint64(13))
}

// TestFlushIndexLastKey verifies that the index records the last key of each
// block, not just the first.
func TestFlushIndexLastKey(t *testing.T) {
	entries := []flushKV{
		{[]byte("apple"), []byte("a")},
		{[]byte("banana"), []byte("b")},
		{[]byte("cherry"), []byte("c")},
	}

	db, sf := newFlushDB(entries)

	if err := db.FlushMemtableToDisk(); err != nil {
		t.Fatal(err)
	}

	written := sf.buf.Bytes()

	// locate the index: footer is last 20 bytes, index precedes it
	footerB := written[len(written)-20:]
	indexLen := binary.LittleEndian.Uint64(footerB[:8])

	indexStart := len(written) - 20 - int(indexLen)
	indexBytes := written[indexStart : len(written)-20]

	// parse the single index entry
	keyLen := binary.LittleEndian.Uint16(indexBytes[:2])
	lastKey := indexBytes[2 : 2+keyLen]

	if !bytes.Equal(lastKey, []byte("cherry")) {
		t.Fatalf("index last key: got %q, want %q", lastKey, "cherry")
	}
}
