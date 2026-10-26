package lifecycle

import (
	"net/mail"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

type DeliveryTarget struct {
	SubjectUserID, ResourceID, RecipientIdentifierID uuid.UUID
	RecipientAddress                                 string
}

func (target DeliveryTarget) Validate() error {
	parsed, err := mail.ParseAddress(target.RecipientAddress)
	if target.SubjectUserID == uuid.Nil || target.ResourceID == uuid.Nil ||
		target.RecipientIdentifierID == uuid.Nil || err != nil || parsed.Name != "" ||
		parsed.Address != target.RecipientAddress {
		return ErrInvalid
	}
	return nil
}

type DeliveryArtifact struct {
	Reference             string
	RecipientIdentifierID uuid.UUID
	Envelope              lifecycledeliveryv1.Envelope
}

func (artifact DeliveryArtifact) Validate(tenantID uuid.UUID,
	purpose CredentialPurpose, token TokenIssue) error {
	reference, err := lifecycledeliveryv1.ParseReference(artifact.Reference)
	if err != nil || artifact.Envelope.Validate(tenantID) != nil ||
		reference.TenantID() != tenantID ||
		reference.ArtifactID() != artifact.Envelope.ArtifactID ||
		artifact.RecipientIdentifierID == uuid.Nil || artifact.Envelope.TokenID != token.ID ||
		string(artifact.Envelope.Purpose) != string(purpose) {
		return ErrInvalid
	}
	return nil
}
