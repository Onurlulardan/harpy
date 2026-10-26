package systemresourcelayout

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

func installationBaselineID(operationID [16]byte, resource ResourceKind) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy:system-resource-layout:installation:v1"))
	_, _ = hash.Write(operationID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(resource))
	var result uuid.UUID
	copy(result[:], hash.Sum(nil))
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}
