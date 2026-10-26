package rbac

import "encoding/hex"

type ID [16]byte

func (id ID) Valid() bool { return id != ID{} }

func (id ID) String() string {
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
