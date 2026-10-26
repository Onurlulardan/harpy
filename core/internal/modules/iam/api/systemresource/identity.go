package systemresource

import "encoding/hex"

type ID [16]byte

func ParseID(value string) (ID, error) {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return ID{}, ErrContract
	}
	compact := make([]byte, 32)
	copy(compact[0:8], value[0:8])
	copy(compact[8:12], value[9:13])
	copy(compact[12:16], value[14:18])
	copy(compact[16:20], value[19:23])
	copy(compact[20:32], value[24:36])
	var id ID
	if _, err := hex.Decode(id[:], compact); err != nil || id == (ID{}) ||
		id.String() != value {
		return ID{}, ErrContract
	}
	return id, nil
}

func (id ID) Valid() bool { return id != (ID{}) }

func (id ID) String() string {
	if !id.Valid() {
		return ""
	}
	var encoded [36]byte
	hex.Encode(encoded[0:8], id[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], id[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], id[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], id[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], id[10:16])
	return string(encoded[:])
}
