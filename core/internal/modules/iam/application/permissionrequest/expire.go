package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ExpireCommand struct {
	OperationID      uuid.UUID
	RequestSHA256    [32]byte
	GroupID          uuid.UUID
	ExpectedRevision int64
	ExpectedState    State
}

func (command ExpireCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.GroupID == uuid.Nil || command.ExpectedRevision < 1 ||
		!CanTransition(command.ExpectedState, StateExpired) {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Expire(
	ctx context.Context,
	scope transaction.TenantScope,
	command ExpireCommand,
) (GroupRef, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return GroupRef{}, ErrInvalid
	}
	result, err := service.store.Expire(ctx, scope, command)
	if err != nil {
		return GroupRef{}, err
	}
	if result.Validate() != nil || result.State != StateExpired {
		return GroupRef{}, ErrUnavailable
	}
	record := TransitionRecord{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		GroupID: result.ID, ActorID: scope.ActorID(),
		Command: TransitionExpire, From: command.ExpectedState,
		To: result.State, Revision: result.Revision,
	}
	if record.Validate() != nil || service.effects.RecordPermissionRequestTransition(
		ctx, scope, record,
	) != nil {
		return GroupRef{}, ErrDecisionEffects
	}
	return result, nil
}
