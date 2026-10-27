package authorization

import (
	"crypto/sha256"
	"crypto/subtle"
)

func sameDigest(left, right [sha256.Size]byte) bool {
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}
