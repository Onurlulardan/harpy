package domain

import (
	"encoding/hex"
	"errors"
)

var ErrInvalidID = errors.New("work identifier is invalid")

type ID [16]byte

func ParseID(value string) (ID, error) {
	var result ID
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return result, ErrInvalidID
	}
	compact := value[:8] + value[9:13] + value[14:18] +
		value[19:23] + value[24:]
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(result) {
		return ID{}, ErrInvalidID
	}
	copy(result[:], decoded)
	if result == (ID{}) {
		return ID{}, ErrInvalidID
	}
	return result, nil
}

func (id ID) Valid() bool { return id != (ID{}) }

func (id ID) UUIDv7() bool {
	return id.Valid() && id[6]>>4 == 7 && id[8]>>6 == 2
}

func (id ID) String() string {
	if !id.Valid() {
		return ""
	}
	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], id[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], id[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], id[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], id[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], id[10:16])
	return string(encoded)
}

func CompareID(left, right ID) int {
	for index := range left {
		if left[index] < right[index] {
			return -1
		}
		if left[index] > right[index] {
			return 1
		}
	}
	return 0
}
