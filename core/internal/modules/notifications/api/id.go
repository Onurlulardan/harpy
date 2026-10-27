package api

import (
	"encoding/hex"
	"errors"
)

var ErrIDInvalid = errors.New("notification identifier is invalid")

type ID [16]byte

func (value ID) Valid() bool { return value != ID{} }

func (value ID) String() string {
	if !value.Valid() {
		return "00000000-0000-0000-0000-000000000000"
	}
	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])
	return string(encoded)
}

func ParseID(text string) (ID, error) {
	if len(text) != 36 || text[8] != '-' || text[13] != '-' ||
		text[18] != '-' || text[23] != '-' {
		return ID{}, ErrIDInvalid
	}
	compact := text[0:8] + text[9:13] + text[14:18] + text[19:23] + text[24:36]
	var value ID
	if _, err := hex.Decode(value[:], []byte(compact)); err != nil || !value.Valid() {
		return ID{}, ErrIDInvalid
	}
	return value, nil
}
