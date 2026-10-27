package initialadminsetup

import (
	"net/mail"
	"strings"
	"time"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"
	"github.com/google/uuid"
)

type PrepareInput struct {
	OperationID, TenantID, UserID, IdentifierID uuid.UUID
	Email                                       string
	CreatedAt                                   time.Time
}

func (input PrepareInput) valid() bool {
	parsed, err := mail.ParseAddress(input.Email)
	return input.OperationID != uuid.Nil && input.TenantID != uuid.Nil &&
		input.UserID != uuid.Nil && input.IdentifierID != uuid.Nil && err == nil &&
		parsed.Name == "" && parsed.Address == input.Email &&
		input.Email == strings.ToLower(input.Email) && len(input.Email) <= 254 &&
		!input.CreatedAt.IsZero() && input.CreatedAt.Location() == time.UTC
}

type Stage struct {
	PrepareInput
	Token     lifecyclecredential.TokenIssue
	Reference string
	Envelope  lifecycledeliveryv1.Envelope
}

func (stage Stage) Validate() error {
	reference, err := lifecycledeliveryv1.ParseReference(stage.Reference)
	if !stage.PrepareInput.valid() || stage.Token.Validate() != nil || err != nil ||
		stage.Envelope.Validate(stage.TenantID) != nil ||
		reference.TenantID() != stage.TenantID ||
		reference.ArtifactID() != stage.Envelope.ArtifactID ||
		stage.Envelope.TokenID != stage.Token.ID ||
		stage.Envelope.SubjectUserID != stage.UserID ||
		stage.Envelope.ResourceID != stage.IdentifierID ||
		stage.Envelope.Purpose != lifecycledeliveryv1.PurposeInitialAdminSetup {
		return ErrInvalid
	}
	return nil
}
