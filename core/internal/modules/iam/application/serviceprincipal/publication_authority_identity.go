package serviceprincipal

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

func PublicationAuthorityStepUpPurpose(
	tenantID, definitionID uuid.UUID,
) ([sha256.Size]byte, error) {
	if tenantID == uuid.Nil || definitionID == uuid.Nil ||
		definitionID.Version() != 7 || definitionID.Variant() != uuid.RFC4122 {
		return [sha256.Size]byte{}, ErrInvalidGrant
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(
		"harpy/workflow/publication-authority/step-up/v1\x00",
	))
	_, _ = hash.Write(tenantID[:])
	_, _ = hash.Write(definitionID[:])
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func publicationAuthorityID(
	definitionID uuid.UUID,
	manifest [sha256.Size]byte,
	purpose string,
) uuid.UUID {
	digest := sha256.New()
	_, _ = digest.Write([]byte("harpy/workflow/publication-authority/id/v1\x00"))
	_, _ = digest.Write(definitionID[:])
	_, _ = digest.Write(manifest[:])
	_, _ = digest.Write([]byte(purpose))
	sum := digest.Sum(nil)
	var result uuid.UUID
	copy(result[:6], definitionID[:6])
	copy(result[6:], sum[:10])
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}

func authorityOperationID(
	request ProvisionPublicationAuthority,
	purpose string,
) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/workflow/publication-authority/operation/v1\x00"))
	_, _ = hash.Write(request.DefinitionID[:])
	_, _ = hash.Write(request.ManifestSHA256[:])
	_, _ = hash.Write([]byte(request.IdempotencyKey))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(purpose))
	sum := sha256.Sum256(hash.Sum(nil))
	return publicationAuthorityID(request.DefinitionID, sum, purpose)
}

func authorityRequestDigest(
	request ProvisionPublicationAuthority,
	aggregateID uuid.UUID,
	purpose string,
) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/workflow/publication-authority/request/v1\x00"))
	_, _ = hash.Write(request.DefinitionID[:])
	_, _ = hash.Write(request.ManifestSHA256[:])
	_, _ = hash.Write(aggregateID[:])
	var versions [16]byte
	binary.BigEndian.PutUint64(versions[:8], uint64(
		request.MutationEvidence.ExpectedTokenVersion,
	))
	binary.BigEndian.PutUint64(versions[8:], uint64(
		request.MutationEvidence.ExpectedAuthorizationVersion,
	))
	_, _ = hash.Write(versions[:])
	_, _ = hash.Write([]byte(request.IdempotencyKey))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(purpose))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}
