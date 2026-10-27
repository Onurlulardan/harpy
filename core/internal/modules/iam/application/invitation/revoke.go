package invitation

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type RevokeCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ID                           uuid.UUID
	ExpectedRevision             int64
	ActorID                      uuid.UUID
	ExpectedAuthorizationVersion int64
}

func (command RevokeCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.ID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.ActorID == uuid.Nil || command.ExpectedAuthorizationVersion < 1 {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Revoke(
	ctx context.Context,
	scope platformtx.TenantScope,
	command RevokeCommand,
) (Ref, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		scope.ActorID() != command.ActorID {
		return Ref{}, ErrInvalid
	}
	result, err := service.store.Revoke(ctx, scope, command)
	if err != nil {
		return Ref{}, err
	}
	if result.Validate() != nil || result.State != StateRevoked {
		return Ref{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if err := service.effects.WriteTenantChange(ctx, scope, AuditChange{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		InvitationID: command.ID, ActorID: command.ActorID,
		CommandKey: "iam.invitation.revoke", ResultState: result.State,
		ResultRevision: result.Revision,
	}); err != nil {
		return Ref{}, ErrEffects
	}
	return result, nil
}
