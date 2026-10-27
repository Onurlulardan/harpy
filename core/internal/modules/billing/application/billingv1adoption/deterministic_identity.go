package billingv1adoptionapp

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

const identityNamespace = "harpy/billing-v1-adoption-identity/v1"

func deterministicIdentity(kind string, source ...string) (uuid.UUID, error) {
	if kind == "" || len(source) == 0 {
		return uuid.Nil, fmt.Errorf("billing V1 adoption identity input is empty")
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(identityNamespace))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(kind))
	for _, value := range source {
		if value == "" {
			return uuid.Nil, fmt.Errorf("billing V1 adoption identity component is empty")
		}
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(value))
	}
	bytes := hash.Sum(nil)[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x70
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	value, err := uuid.FromBytes(bytes)
	if err != nil || value == uuid.Nil || value.Version() != 7 || value.Variant() != uuid.RFC4122 {
		return uuid.Nil, fmt.Errorf("billing V1 adoption identity is invalid")
	}
	return value, nil
}

func mustIdentity(kind string, source ...string) string {
	value, err := deterministicIdentity(kind, source...)
	if err != nil {
		panic(err)
	}
	return value.String()
}
