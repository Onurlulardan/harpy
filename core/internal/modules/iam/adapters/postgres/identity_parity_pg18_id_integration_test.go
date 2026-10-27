//go:build integration

package postgres

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

func iamInstallationDerivedID(operation, label string) string {
	operationID := uuid.MustParse(operation)
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy:iam:installation:v1"))
	_, _ = hash.Write(operationID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(label))
	result := [16]byte(hash.Sum(nil)[:16])
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return uuid.UUID(result).String()
}
