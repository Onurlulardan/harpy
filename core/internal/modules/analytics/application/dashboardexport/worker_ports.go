package dashboardexport

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkerStore interface {
	MarkEffectStarted(context.Context, EffectStart) error
	CompleteReady(context.Context, ReadyWrite) (State, error)
	MarkOutcomeUnknown(context.Context, RenderUnknown) error
	FinishRender(context.Context, RenderFailure) (State, error)
	ControlExpiry(
		context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time,
	) (ExpiryControl, error)
	ClaimDelete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (DeleteClaim, error)
	CompleteDeleted(context.Context, DeleteClaim, DeleteReceipt, time.Time) error
}

type AuthorizationRequest struct {
	TenantID, ActorID, DashboardID                         uuid.UUID
	TokenVersion, AuthorizationVersion, AuthorizationClock int64
	Permission                                             string
}

type WorkerAuthorizer interface {
	Authorize(context.Context, AuthorizationRequest) error
}
