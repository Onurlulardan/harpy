package customerprovisioning

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

func operationID(environmentID uuid.UUID, idempotencyKey string) uuid.UUID {
	return stableID(environmentID, "operation\x00"+idempotencyKey)
}

func stableID(parent uuid.UUID, label string) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/tenancy/customer-provision/v1\x00"))
	_, _ = hash.Write(parent[:])
	_, _ = hash.Write([]byte(label))
	var result uuid.UUID
	copy(result[:], hash.Sum(nil))
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}
