package publication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) transitionStore(
	ctx context.Context, scope platformtx.TenantScope, operation string,
	releaseID uuid.UUID, approval bool,
) (Store, ScopeIdentity, Snapshot, timeBinding, error) {
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		operation, PermissionRead)
	if err != nil {
		return nil, ScopeIdentity{}, Snapshot{}, timeBinding{}, err
	}
	current, err := store.Get(ctx, releaseID)
	if err != nil || validateSnapshot(current) != nil {
		return nil, ScopeIdentity{}, Snapshot{}, timeBinding{}, ErrConflict
	}
	permission, ok := permissionForAudience(current.Audience, approval)
	if !ok || application.authorizer.Authorize(ctx, scope, permission) != nil {
		return nil, ScopeIdentity{}, Snapshot{}, timeBinding{}, ErrForbidden
	}
	timing, err := application.freshTransitionTime(ctx, scope)
	if err != nil {
		return nil, ScopeIdentity{}, Snapshot{}, timeBinding{}, err
	}
	return store, identity, current, timing, nil
}

func (application *Application) freshTransitionTime(
	ctx context.Context, scope platformtx.TenantScope,
) (timeBinding, error) {
	now := application.clock.Now()
	stepUp, err := application.authorizer.FreshStepUp(ctx, scope, now)
	if err != nil || !validTime(now) || !validTime(stepUp) || stepUp.After(now) ||
		now.Sub(stepUp) > 10*time.Minute {
		return timeBinding{}, ErrForbidden
	}
	return timeBinding{now: now, stepUp: stepUp}, nil
}

type timeBinding struct{ now, stepUp time.Time }
