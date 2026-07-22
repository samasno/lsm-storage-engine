package db

import "encoding/binary"

// Sequence number and action are always greater than 0
func extractKeyRaw(encodedKey []byte) []byte {
	if nil == encodedKey || 9 > len(encodedKey) {
		return nil
	}

	elen := len(encodedKey)
	rawKey := encodedKey[:elen-8]

	return rawKey
}

func extractKeySequenceNumber(encodedKey []byte) uint64 {
	if nil == encodedKey || 9 > len(encodedKey) {
		return 0
	}

	sequenceCode := binary.LittleEndian.Uint64(encodedKey[len(encodedKey)-8:])

	return sequenceCode >> 8
}

func extractKeyAction(encodedKey []byte) Action {
	if nil == encodedKey || 9 > len(encodedKey) {
		return 0
	}

	action := binary.LittleEndian.Uint64(encodedKey[len(encodedKey)-8:])

	return Action(action)
}

// used to enforce invariants at in package level
func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
