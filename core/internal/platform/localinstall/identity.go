package localinstall

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const identityNamespace = "harpy-v2/local-installation/v1"

type Identity struct {
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	RefreshOperationID     uuid.UUID
	OfferingOperationID    uuid.UUID
	OfferingVersionID      uuid.UUID
}

func DeriveIdentity(environmentID uuid.UUID) Identity {
	return Identity{
		LicenseID:              stableV7(environmentID, "license"),
		LicenseeOrganizationID: stableV7(environmentID, "licensee-organization"),
		RefreshOperationID:     stableV7(environmentID, "license-refresh-sequence-1"),
		OfferingOperationID:    stableV7(environmentID, "license-offering-sequence-1"),
		OfferingVersionID:      stableV7(environmentID, "license-offering-version-1"),
	}
}

func NextCredentialID(operationID uuid.UUID, version int64) uuid.UUID {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(version))
	name := append(append([]byte("credential"), operationID[:]...), encoded[:]...)
	return uuid.NewSHA1(uuid.MustParse("aaec60b3-e3e4-5f96-a52a-30bb1f5f4663"), name)
}

func stableV7(environmentID uuid.UUID, purpose string) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte(identityNamespace))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(environmentID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(purpose))
	value := uuid.UUID(hash.Sum(nil)[:16])
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	return value
}
