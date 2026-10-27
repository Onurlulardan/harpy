package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CancelCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	GroupID                      uuid.UUID
	ExpectedRevision             int64
	ExpectedState                State
	RequesterID                  uuid.UUID
	ExpectedAuthorizationVersion int64
}

func (command CancelCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.GroupID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.RequesterID == uuid.Nil || command.ExpectedAuthorizationVersion < 1 ||
		!CanTransition(command.ExpectedState, StateCancelled) {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Cancel(
	ctx context.Context,
	scope transaction.TenantScope,
	command CancelCommand,
) (GroupRef, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		scope.ActorID() != command.RequesterID {
		return GroupRef{}, ErrInvalid
	}
	result, err := service.store.Cancel(ctx, scope, command)
	if err != nil {
		return GroupRef{}, err
	}
	if result.Validate() != nil || result.State != StateCancelled {
		return GroupRef{}, ErrUnavailable
	}
	record := TransitionRecord{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		GroupID: result.ID, ActorID: command.RequesterID,
		Command: TransitionCancel, From: command.ExpectedState,
		To: result.State, Revision: result.Revision,
	}
	if record.Validate() != nil || service.effects.RecordPermissionRequestTransition(
		ctx, scope, record,
	) != nil {
		return GroupRef{}, ErrDecisionEffects
	}
	return result, nil
}
