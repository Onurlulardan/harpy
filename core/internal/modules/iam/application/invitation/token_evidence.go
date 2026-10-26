package invitation

import (
	"encoding/base64"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

type TokenIssue struct {
	ID                  uuid.UUID
	Selector            [16]byte
	HashKeyID           string
	SecretHMAC          [32]byte
	BindingSHA256       [32]byte
	DeliveryArtifactRef string
	DeliveryEnvelope    deliveryartifactv1.Envelope
}

func (issue TokenIssue) Validate() error {
	reference, referenceErr := deliveryartifactv1.ParseReference(
		issue.DeliveryArtifactRef,
	)
	if issue.ID == uuid.Nil || issue.Selector == [16]byte{} ||
		!validTokenKeyID(issue.HashKeyID) ||
		issue.SecretHMAC == [32]byte{} || issue.BindingSHA256 == [32]byte{} ||
		referenceErr != nil || reference.ArtifactID() != issue.DeliveryEnvelope.ArtifactID ||
		issue.DeliveryEnvelope.TokenID != issue.ID ||
		issue.DeliveryEnvelope.KeyID == "" {
		return ErrInvalid
	}
	return nil
}

func (issue TokenIssue) MatchesTenant(tenantID uuid.UUID) error {
	reference, err := deliveryartifactv1.ParseReference(issue.DeliveryArtifactRef)
	if issue.Validate() != nil || err != nil || reference.TenantID() != tenantID ||
		issue.DeliveryEnvelope.Validate(tenantID) != nil {
		return ErrInvalid
	}
	return nil
}

func validTokenKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}
