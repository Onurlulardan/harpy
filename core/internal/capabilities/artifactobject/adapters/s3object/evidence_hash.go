package s3object

import (
	"crypto/sha256"
	"encoding/binary"
)

func evidenceHash(contract string, fields ...string) [sha256.Size]byte {
	hash := sha256.New()
	writeEvidenceField(hash, contract)
	for _, field := range fields {
		writeEvidenceField(hash, field)
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type evidenceWriter interface{ Write([]byte) (int, error) }

func writeEvidenceField(destination evidenceWriter, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = destination.Write(size[:])
	_, _ = destination.Write([]byte(value))
}
