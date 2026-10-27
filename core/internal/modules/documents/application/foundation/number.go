package foundation

import "encoding/binary"

func Int64Bytes(value int64) []byte {
	var result [8]byte
	binary.BigEndian.PutUint64(result[:], uint64(value))
	return result[:]
}
