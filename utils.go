package main

import (
	"encoding/binary"
)

// itob converts an int to a byte slice for BoltDB keys
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi converts a byte slice to an int for BoltDB keys
func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}
