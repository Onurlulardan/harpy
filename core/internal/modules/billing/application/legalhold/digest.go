package legalhold

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

func requestDigest(values ...string) []byte {
	hash := sha256.New()
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	return hash.Sum(nil)
}

func idText(value uuid.UUID) string { return value.String() }
