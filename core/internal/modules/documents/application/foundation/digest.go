package foundation

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

func Digest(parts ...[]byte) [32]byte {
	digest := sha256.New()
	for _, part := range parts {
		writeDigestPart(digest, part)
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeDigestPart(digest hash.Hash, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write(value)
}
