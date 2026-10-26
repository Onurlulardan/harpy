package dashboard

import (
	"context"
	"encoding/binary"
	"fmt"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type DeleteRequest struct {
	DashboardID          uuid.UUID
	ExpectedRootRevision uint64
	IdempotencyKey       string
}

func (application *Application) Delete(
	ctx context.Context, scope platformtx.TenantScope, request DeleteRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.DashboardID) || request.ExpectedRootRevision == 0 ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationDelete, PermissionDelete)
	if err != nil {
		return Snapshot{}, err
	}
	return application.deleteBound(ctx, scope, store, identity, request)
}

func (application *Application) deleteBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request DeleteRequest,
) (Snapshot, error) {
	ids, err := application.nextIDs(1)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	revision := make([]byte, 8)
	binary.BigEndian.PutUint64(revision, request.ExpectedRootRevision)
	requestDigest := digest([]byte(OperationDelete), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.DashboardID[:], revision)
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceDelete, ids[0], request.DashboardID,
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Delete(ctx, DeleteMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, DashboardID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest,
		ExpectedRootRevision: request.ExpectedRootRevision, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("delete analytics dashboard: %w", err)
	}
	if validateSnapshot(result) != nil || result.DashboardID != decision.OwnerFactID ||
		result.Lifecycle != analyticsdashboard.LifecycleRetired ||
		replay != decision.Replay || (!replay &&
		result.RootRevision != request.ExpectedRootRevision+1) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
