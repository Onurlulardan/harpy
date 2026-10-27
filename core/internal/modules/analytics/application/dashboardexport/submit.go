package dashboardexport

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Submit(
	ctx context.Context, scope platformtx.TenantScope, request SubmitRequest,
) (Snapshot, error) {
	parameters, err := canonicalParameters(request.Parameters)
	if !validUUIDv7(request.DashboardID) || !request.Format.Valid() ||
		!idempotencyPattern.MatchString(request.IdempotencyKey) || err != nil ||
		request.Auth.TokenVersion < 1 || request.Auth.AuthorizationVersion < 1 ||
		request.Auth.AuthorizationClock < 1 {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, OperationCreate)
	if err != nil {
		return Snapshot{}, err
	}
	return application.submitBound(ctx, scope, store, identity, request, parameters)
}

func (application *Application) submitBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request SubmitRequest, parameters []byte,
) (Snapshot, error) {
	ids, err := application.nextIDs(4)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	digest := requestDigest(OperationCreate, identity.TenantID, identity.ActorID,
		request.DashboardID[:], []byte(request.Format), parameters)
	mutation := SubmitMutation{OperationID: ids[0], ExportID: ids[1], JobID: ids[2],
		ExpiryJobID: ids[3],
		ActorID:     identity.ActorID, DashboardID: request.DashboardID, Format: request.Format,
		Parameters: parameters, IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		Auth: request.Auth, MaximumRows: MaximumRows, MaximumPlaintextBytes: MaximumPlaintextBytes,
		OccurredAt: now, ExpiresAt: now.Add(ExportLifetime)}
	result, replay, err := store.Submit(ctx, mutation)
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result, request.DashboardID, result.ExportID) != nil ||
		result.State != StateQueued {
		return Snapshot{}, ErrUnavailable
	}
	if !replay && application.scheduler.Schedule(ctx, scope, Job{JobID: result.JobID,
		ExportID: result.ExportID, Action: JobRender, RunAt: now}) != nil {
		return Snapshot{}, fmt.Errorf("schedule analytics dashboard export: %w", ErrUnavailable)
	}
	if !replay && application.scheduler.Schedule(ctx, scope, Job{JobID: mutation.ExpiryJobID,
		ExportID: result.ExportID, Action: JobExpire, RunAt: mutation.ExpiresAt}) != nil {
		return Snapshot{}, fmt.Errorf("schedule analytics dashboard export expiry: %w",
			ErrUnavailable)
	}
	return result, nil
}

func (application *Application) nextIDs(count int) ([]uuid.UUID, error) {
	values := make([]uuid.UUID, count)
	for index := range values {
		value, err := application.ids.NewUUIDv7()
		if err != nil || !validUUIDv7(value) {
			return nil, ErrUnavailable
		}
		values[index] = value
	}
	return values, nil
}
