package campaign

import "bytes"

type Identifier [16]byte

func validIdentifier(identifier Identifier) bool {
	return identifier != Identifier{}
}

func compareIdentifier(left, right Identifier) int {
	return bytes.Compare(left[:], right[:])
}
