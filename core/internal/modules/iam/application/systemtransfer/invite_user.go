package systemtransfer

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (worker *Worker) inviteUser(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	row SourceRow,
	values publicresource.FieldValues,
	key string,
) (RowReceipt, error) {
	input, err := invitationInput(values, key)
	if err != nil {
		return rejectedReceipt(row, "INVITATION_FIELDS_INVALID")
	}
	result, err := worker.invitations.Invite(ctx, scope, evidence, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvitationInvalid):
			return rejectedReceipt(row, "INVITATION_FIELDS_INVALID")
		case errors.Is(err, ErrInvitationConflict):
			return rejectedReceipt(row, "DUPLICATE_CONFLICT")
		case errors.Is(err, ErrInvitationDenied):
			return rejectedReceipt(row, "COMMAND_FORBIDDEN")
		default:
			return RowReceipt{}, ErrInvitationUnavailable
		}
	}
	if result.Validate() != nil {
		return RowReceipt{}, ErrInvitationUnavailable
	}
	return completedReceipt(row, "created", "invitation", uuid.UUID(result.ID),
		result.Revision)
}
