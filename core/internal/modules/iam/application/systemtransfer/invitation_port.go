package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type InvitationInput struct {
	Email          string
	RoleID         publicresource.ID
	IdempotencyKey string
}

func (input InvitationInput) Validate() error {
	if input.Email == "" || !input.RoleID.Valid() || input.IdempotencyKey == "" {
		return ErrInvitationInvalid
	}
	return nil
}

type InvitationResult struct {
	ID publicresource.ID
	// Revision is the canonical command receipt revision, including on replay.
	Revision int64
	Replayed bool
}

func (result InvitationResult) Validate() error {
	if !result.ID.Valid() || result.Revision < 1 {
		return ErrInvitationInvalid
	}
	return nil
}

// InvitationCommand is intentionally password-free. Implementations must resolve
// and authorize the allowed role through IAM's canonical invitation application.
type InvitationCommand interface {
	Invite(
		context.Context,
		platformtx.TenantScope,
		publicresource.ActorEvidence,
		InvitationInput,
	) (InvitationResult, error)
}
