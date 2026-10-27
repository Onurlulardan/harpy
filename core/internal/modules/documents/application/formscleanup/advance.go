package formscleanup

import (
	"context"
	"encoding/json"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Advance(ctx context.Context,
	scope platformtx.TenantScope, command formscleanupapi.Advance,
) (formscleanupapi.Progress, error) {
	operationID, err := foundation.ParseID(command.OperationID)
	if application == nil || err != nil || !operationID.UUIDv7() {
		return formscleanupapi.Progress{}, ErrInvalid
	}
	store, err := application.bind(ctx, scope)
	if err != nil {
		return formscleanupapi.Progress{}, err
	}
	return application.advance(ctx, scope, store, operationID)
}

func (application *Application) advance(ctx context.Context,
	scope platformtx.TenantScope, store Store, operationID domain.ID,
) (formscleanupapi.Progress, error) {
	now := application.clock.Now().UTC().Truncate(1e3)
	snapshot, candidates, err := store.CaptureLinks(ctx, operationID, now)
	if err != nil {
		return formscleanupapi.Progress{}, mapError(err)
	}
	for _, candidate := range candidates {
		if !candidate.ItemID.UUIDv7() || candidate.PurgeRunAt.IsZero() {
			return formscleanupapi.Progress{}, ErrUnavailable
		}
		if !candidate.JobID.Valid() {
			jobID, idErr := application.ids.NewUUIDv7()
			if idErr != nil || !jobID.UUIDv7() {
				return formscleanupapi.Progress{}, ErrUnavailable
			}
			candidate, err = store.StageEnqueue(ctx, EnqueuePlan{
				OperationID: operationID, ItemID: candidate.ItemID,
				JobID: jobID, RunAt: candidate.PurgeRunAt,
			})
			if err != nil {
				return formscleanupapi.Progress{}, mapError(err)
			}
		}
		payload, encodeErr := json.Marshal(map[string]string{
			"item_id": candidate.ItemID.String(),
		})
		if encodeErr != nil || !candidate.JobID.UUIDv7() ||
			!candidate.ItemID.UUIDv7() || candidate.PurgeRunAt.IsZero() {
			return formscleanupapi.Progress{}, ErrUnavailable
		}
		if err = application.jobs.EnqueuePurge(ctx, scope, candidate.JobID,
			candidate.ItemID, candidate.PurgeRunAt, payload); err != nil {
			return formscleanupapi.Progress{}, ErrUnavailable
		}
		if err = store.AcknowledgeEnqueue(ctx, Enqueued{OperationID: operationID,
			ItemID: candidate.ItemID, JobID: candidate.JobID, RunAt: candidate.PurgeRunAt,
			EnqueuedAt: now}); err != nil {
			return formscleanupapi.Progress{}, mapError(err)
		}
	}
	snapshot, err = store.SealReceipt(ctx, operationID, now)
	if err != nil {
		return formscleanupapi.Progress{}, mapError(err)
	}
	return exportProgress(snapshot)
}

func (application *Application) bind(ctx context.Context,
	scope platformtx.TenantScope,
) (Store, error) {
	if ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, ErrUnavailable
	}
	store, err := application.binder.Bind(ctx, scope)
	if err != nil || foundation.NilPort(store) {
		return nil, ErrUnavailable
	}
	return store, nil
}
