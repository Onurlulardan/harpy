package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Decision string

const (
	DecisionApprove Decision = "APPROVE"
	DecisionReject  Decision = "REJECT"
)

type DecideCommand struct {
	OperationID                          uuid.UUID
	RequestSHA256                        [32]byte
	GroupID                              uuid.UUID
	ExpectedRevision                     int64
	ExpectedState                        State
	ReviewerID                           uuid.UUID
	ExpectedReviewerTokenVersion         int64
	ExpectedReviewerAuthorizationVersion int64
	ReviewTaskID                         uuid.UUID
	Decision                             Decision
	StepUp                               StepUpEvidence
}

type DecisionResult struct {
	Code               string
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

type DecisionRecord struct {
	OperationID   uuid.UUID
	RequestSHA256 [32]byte
	GroupID       uuid.UUID
	ReviewerID    uuid.UUID
	ReviewTaskID  uuid.UUID
	Decision      Decision
	Result        DecisionResult
}

func (command DecideCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.GroupID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.ExpectedState != StateInReview || command.ReviewerID == uuid.Nil ||
		command.ExpectedReviewerTokenVersion < 1 ||
		command.ExpectedReviewerAuthorizationVersion < 1 ||
		command.ReviewTaskID == uuid.Nil || command.StepUp.Validate() != nil ||
		(command.Decision != DecisionApprove && command.Decision != DecisionReject) {
		return ErrInvalid
	}
	target := StateApproved
	if command.Decision == DecisionReject {
		target = StateRejected
	}
	if !CanTransition(command.ExpectedState, target) {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Decide(
	ctx context.Context,
	scope transaction.TenantScope,
	command DecideCommand,
) (DecisionResult, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil || scope.ActorID() != command.ReviewerID {
		return DecisionResult{}, ErrInvalid
	}
	result, err := service.store.Decide(ctx, scope, command)
	if err != nil {
		return DecisionResult{}, err
	}
	if result.Revision < 1 || result.AuthorizationClock < 1 || result.Code == "" {
		return DecisionResult{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	record := DecisionRecord{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		GroupID:    command.GroupID,
		ReviewerID: command.ReviewerID, ReviewTaskID: command.ReviewTaskID,
		Decision: command.Decision, Result: result,
	}
	if record.Validate() != nil {
		return DecisionResult{}, ErrUnavailable
	}
	if err := service.effects.RecordPermissionRequestDecision(
		ctx, scope, record,
	); err != nil {
		return DecisionResult{}, ErrDecisionEffects
	}
	return result, nil
}
