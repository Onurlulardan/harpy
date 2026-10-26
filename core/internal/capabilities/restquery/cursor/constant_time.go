package cursor

import (
	"crypto/subtle"
	"encoding/binary"
)

func constantTimeBytes(left, right []byte) int {
	return subtle.ConstantTimeCompare(left, right)
}

func constantTimeString(left, right string) int {
	return constantTimeBytes([]byte(left), []byte(right))
}

func constantTimeInt(left, right int) int {
	var leftBytes, rightBytes [8]byte
	binary.BigEndian.PutUint64(leftBytes[:], uint64(left))
	binary.BigEndian.PutUint64(rightBytes[:], uint64(right))
	return constantTimeBytes(leftBytes[:], rightBytes[:])
}

func constantTimeSorts(left []decodedSort, right []SortDefinition) int {
	if len(left) != len(right) {
		return 0
	}
	match := 1
	for index := range left {
		match &= constantTimeBytes(left[index].fieldID[:], right[index].fieldID[:])
		match &= subtle.ConstantTimeByteEq(
			uint8(left[index].direction), uint8(right[index].direction),
		)
		match &= subtle.ConstantTimeByteEq(
			uint8(left[index].logicalType), uint8(right[index].logicalType),
		)
	}
	return match
}
