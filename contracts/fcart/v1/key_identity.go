package fcartv1

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

type KeyIdentity struct {
	TenantID          uuid.UUID
	ArtifactID        uuid.UUID
	AttemptGeneration uint64
}

func (identity KeyIdentity) Validate() error {
	if identity.TenantID == uuid.Nil || identity.ArtifactID == uuid.Nil ||
		identity.AttemptGeneration == 0 {
		return fmt.Errorf("FCART1 key identity is invalid")
	}
	return nil
}

func (identity KeyIdentity) aad(keyID string) ([]byte, error) {
	if err := identity.Validate(); err != nil || !validKeyID(keyID) {
		return nil, ErrInvalidKeyEnvelope
	}
	aad := make([]byte, 0, len(Version)+16+16+8+2+len(keyID))
	aad = append(aad, Version...)
	aad = append(aad, identity.TenantID[:]...)
	aad = append(aad, identity.ArtifactID[:]...)
	var integer [8]byte
	binary.BigEndian.PutUint64(integer[:], identity.AttemptGeneration)
	aad = append(aad, integer[:]...)
	aad = append(aad, byte(len(keyID)>>8), byte(len(keyID)))
	return append(aad, keyID...), nil
}
