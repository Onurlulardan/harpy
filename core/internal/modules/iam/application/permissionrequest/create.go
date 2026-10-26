package permissionrequest

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CreateItem struct {
	ID               uuid.UUID
	Ordinal          int16
	Target           Target
	PermissionID     uuid.UUID
	PlannedBindingID uuid.UUID
}

type CreateCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	GroupID                      uuid.UUID
	RequesterID                  uuid.UUID
	Mode                         Mode
	ExpectedAuthorizationVersion int64
	StepUpPurposeSHA256          [32]byte
	ExpiresAt                    time.Time
	Requests                     []CreateItem
}

func (command CreateCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.GroupID == uuid.Nil || command.RequesterID == uuid.Nil ||
		(command.Mode != ModeSelfService && command.Mode != ModePrivileged) ||
		command.ExpectedAuthorizationVersion < 1 ||
		command.StepUpPurposeSHA256 == [32]byte{} || command.ExpiresAt.IsZero() ||
		len(command.Requests) < 1 || len(command.Requests) > 100 {
		return ErrInvalid
	}
	requestIDs := make(map[uuid.UUID]struct{}, len(command.Requests))
	bindingIDs := make(map[uuid.UUID]struct{}, len(command.Requests))
	for index, item := range command.Requests {
		if item.ID == uuid.Nil || item.Ordinal != int16(index+1) ||
			item.Target.Validate() != nil || item.PermissionID == uuid.Nil ||
			item.PlannedBindingID == uuid.Nil {
			return ErrInvalid
		}
		if command.Mode == ModeSelfService &&
			(item.Target.Kind != TargetUser || item.Target.ID != command.RequesterID) {
			return ErrDenied
		}
		if _, exists := requestIDs[item.ID]; exists {
			return ErrInvalid
		}
		if _, exists := bindingIDs[item.PlannedBindingID]; exists {
			return ErrInvalid
		}
		requestIDs[item.ID] = struct{}{}
		bindingIDs[item.PlannedBindingID] = struct{}{}
	}
	return nil
}

func (service *Service) Create(
	ctx context.Context,
	scope transaction.TenantScope,
	command CreateCommand,
) (GroupRef, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		scope.ActorID() != command.RequesterID {
		return GroupRef{}, ErrInvalid
	}
	result, err := service.store.Create(ctx, scope, command)
	if err != nil {
		return GroupRef{}, err
	}
	if result.Validate() != nil || result.State != StatePending {
		return GroupRef{}, ErrUnavailable
	}
	record := TransitionRecord{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		GroupID: result.ID, ActorID: command.RequesterID,
		Command: TransitionCreate, To: result.State, Revision: result.Revision,
	}
	if record.Validate() != nil || service.effects.RecordPermissionRequestTransition(
		ctx, scope, record,
	) != nil {
		return GroupRef{}, ErrDecisionEffects
	}
	return result, nil
}
