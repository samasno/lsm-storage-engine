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
