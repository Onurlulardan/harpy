package dashboardexport

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Delete(
	ctx context.Context, scope platformtx.TenantScope, request DeleteRequest,
) (Snapshot, error) {
	if !validUUIDv7(request.DashboardID) || !validUUIDv7(request.ExportID) ||
		!idempotencyPattern.MatchString(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, OperationDelete)
	if err != nil {
		return Snapshot{}, err
	}
	ids, err := application.nextIDs(2)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	digest := requestDigest(OperationDelete, identity.TenantID, identity.ActorID,
		request.DashboardID[:], request.ExportID[:])
	result, replay, schedule, err := store.RequestDelete(ctx, DeleteMutation{
		OperationID: ids[0], DeleteJobID: ids[1], ActorID: identity.ActorID,
		DeleteRequest: request, RequestSHA256: digest, OccurredAt: now})
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result, request.DashboardID, request.ExportID) != nil {
		return Snapshot{}, ErrUnavailable
	}
	if !replay && schedule && application.scheduler.Schedule(ctx, scope, Job{
		JobID: ids[1], ExportID: request.ExportID, Action: JobDelete, RunAt: now}) != nil {
		return Snapshot{}, fmt.Errorf("schedule analytics dashboard export deletion: %w", ErrUnavailable)
	}
	return result, nil
}
