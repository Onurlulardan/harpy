package adoption

import (
	"crypto/sha256"
	"encoding/binary"
)

func sourceDigest(value LegacyRecord) [32]byte {
	document := make([]byte, 0, 160)
	document = append(document, value.SourceID[:]...)
	document = append(document, value.EnvironmentID[:]...)
	document = append(document, value.TenantID[:]...)
	document = append(document, value.ActorID[:]...)
	document = appendString(document, value.Method)
	document = appendString(document, value.Path)
	var numbers [24]byte
	binary.BigEndian.PutUint64(numbers[0:8], uint64(value.StatusCode))
	binary.BigEndian.PutUint64(numbers[8:16], uint64(value.DurationMicros))
	binary.BigEndian.PutUint64(numbers[16:24], uint64(value.OccurredAt.UTC().UnixMicro()))
	document = append(document, numbers[:]...)
	return sha256.Sum256(document)
}

func appendString(document []byte, value string) []byte {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	document = append(document, size[:]...)
	return append(document, value...)
}
