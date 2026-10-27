package query

import (
	"context"
	"encoding/binary"
	"fmt"

	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type DeleteRequest struct {
	QueryID              uuid.UUID
	ExpectedRootRevision uint64
	IdempotencyKey       string
}

func (application *Application) Delete(
	ctx context.Context, scope platformtx.TenantScope, request DeleteRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.QueryID) || request.ExpectedRootRevision == 0 ||
		!validIdempotency(request.IdempotencyKey) {
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
	operationID, err := application.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(operationID) {
		return Snapshot{}, ErrUnavailable
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	revision := make([]byte, 8)
	binary.BigEndian.PutUint64(revision, request.ExpectedRootRevision)
	digest := requestDigest([]byte(OperationDelete), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.QueryID[:], revision)
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceDelete, operationID, request.QueryID,
		request.IdempotencyKey, digest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Delete(ctx, DeleteMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, QueryID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		ExpectedRootRevision: request.ExpectedRootRevision, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("delete analytics query: %w", err)
	}
	if validateSnapshot(result) != nil || result.QueryID != decision.OwnerFactID ||
		result.Lifecycle != analyticsquery.LifecycleRetired || (!replay &&
		result.RootRevision != request.ExpectedRootRevision+1) || replay != decision.Replay {
		return Snapshot{}, ErrUnavailable
	}
	application.invalidate(ctx, request.QueryID)
	return result, nil
}
