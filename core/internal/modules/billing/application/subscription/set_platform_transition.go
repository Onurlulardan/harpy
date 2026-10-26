package subscriptionapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type PlatformTransitionRequest struct {
	ActorID           uuid.UUID
	SubscriptionID    uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
	TargetState       string
	Reason            string
}

func (commands *Commands) SetPlatformTransition(
	ctx context.Context,
	request PlatformTransitionRequest,
) (Result, error) {
	now, err := commands.instant()
	reason, reasonValid := canonicalReason(request.Reason)
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.ActorID) ||
		!validID(request.SubscriptionID) || !validV7(request.SourceOperationID) ||
		request.ExpectedRevision < 1 || !subscriptiondomain.State(request.TargetState).Valid() ||
		!reasonValid {
		return Result{}, ErrRequest
	}
	execution := ActorExecution{
		Execution: Execution{
			OperationID: OperationPlatformTransition, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision,
			RequestSHA256:     platformTransitionDigest(request, reason),
		},
		ActorID: request.ActorID, Permission: permissionPlatformManage,
	}
	result, err := commands.store.Actor(ctx, execution, func(locked LockedState) (CommitPlan, error) {
		return commands.planPlatformTransition(locked, execution, request, reason, now)
	})
	if err != nil {
		return Result{}, fmt.Errorf("transition platform billing subscription: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func platformTransitionDigest(
	request PlatformTransitionRequest,
	reason string,
) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationPlatformTransition)
	writer.identifier(request.ActorID)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.SourceOperationID)
	writer.integer(request.ExpectedRevision)
	writer.text(request.TargetState)
	writer.text(reason)
	return writer.sum()
}
