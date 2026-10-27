package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ReviewCommand struct {
	OperationID                          uuid.UUID
	RequestSHA256                        [32]byte
	GroupID                              uuid.UUID
	ExpectedRevision                     int64
	ExpectedState                        State
	ReviewerID                           uuid.UUID
	ExpectedReviewerAuthorizationVersion int64
	ReviewTaskID                         uuid.UUID
}

func (command ReviewCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.GroupID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.ExpectedState != StatePending || command.ReviewerID == uuid.Nil ||
		command.ExpectedReviewerAuthorizationVersion < 1 ||
		command.ReviewTaskID == uuid.Nil ||
		!CanTransition(command.ExpectedState, StateInReview) {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Review(
	ctx context.Context,
	scope transaction.TenantScope,
	command ReviewCommand,
) (GroupRef, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		scope.ActorID() != command.ReviewerID {
		return GroupRef{}, ErrInvalid
	}
	result, err := service.store.Review(ctx, scope, command)
	if err != nil {
		return GroupRef{}, err
	}
	if result.Validate() != nil || result.State != StateInReview {
		return GroupRef{}, ErrUnavailable
	}
	record := TransitionRecord{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		GroupID: result.ID, ActorID: command.ReviewerID,
		ReviewTaskID: command.ReviewTaskID, Command: TransitionReview,
		From: command.ExpectedState, To: result.State, Revision: result.Revision,
	}
	if record.Validate() != nil || service.effects.RecordPermissionRequestTransition(
		ctx, scope, record,
	) != nil {
		return GroupRef{}, ErrDecisionEffects
	}
	return result, nil
}
