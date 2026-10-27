package dashboard

import (
	"context"
	"regexp"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var surfacePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

type SetDefaultRequest struct {
	DashboardID    uuid.UUID
	Surface        string
	IdempotencyKey string
}

func (application *Application) SetDefault(
	ctx context.Context, scope platformtx.TenantScope, request SetDefaultRequest,
) error {
	if !validUUIDv7(request.DashboardID) || !surfacePattern.MatchString(request.Surface) ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationSetDefault, PermissionRead)
	if err != nil {
		return err
	}
	return application.setDefaultBound(ctx, scope, store, identity, request)
}

func (application *Application) setDefaultBound(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	request SetDefaultRequest,
) error {
	ids, err := application.nextIDs(1)
	if err != nil {
		return err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return ErrUnavailable
	}
	requestDigest := digest([]byte(OperationSetDefault), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.DashboardID[:], []byte(request.Surface))
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceUpdate, ids[0], request.DashboardID,
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return err
	}
	replay, err := store.SetDefault(ctx, DefaultMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, PrincipalID: identity.ActorID,
		DashboardID: decision.OwnerFactID, Surface: request.Surface,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest, OccurredAt: now})
	if err != nil {
		return err
	}
	if replay != decision.Replay {
		return ErrUnavailable
	}
	return nil
}
