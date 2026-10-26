package main

import (
	"hash/crc32"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

func checksum(value []byte) *wrapperspb.Int64Value {
	return wrapperspb.Int64(int64(crc32.Checksum(value, castagnoliTable)))
}

func validChecksum(value []byte, expected *wrapperspb.Int64Value) bool {
	return expected != nil && expected.Value >= 0 &&
		uint32(expected.Value) == crc32.Checksum(value, castagnoliTable)
}
