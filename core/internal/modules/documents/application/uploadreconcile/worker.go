package uploadreconcile

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Worker struct {
	binder    StoreBinder
	inventory Inventory
	billing   Billing
	jobs      Jobs
	ids       foundation.IDSource
	clock     foundation.Clock
}

func NewWorker(binder StoreBinder, inventory Inventory, billing Billing, jobs Jobs,
	ids foundation.IDSource, clock foundation.Clock,
) (*Worker, error) {
	for _, dependency := range []any{binder, inventory, billing, jobs, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents upload reconciliation dependencies are invalid")
		}
	}
	return &Worker{binder: binder, inventory: inventory, billing: billing,
		jobs: jobs, ids: ids, clock: clock}, nil
}

func (worker *Worker) Run(ctx context.Context, runner JobRunner, tenantID,
	uploadID domain.ID, generation uint64,
) (Result, error) {
	if worker == nil || foundation.NilPort(runner) || !tenantID.Valid() ||
		!uploadID.UUIDv7() || generation == 0 || !platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	plan, err := worker.prepare(ctx, runner, tenantID, uploadID, generation)
	if err != nil {
		return Result{}, err
	}
	if !plan.Actionable {
		return Result{UploadID: uploadID, State: plan.State, Noop: true}, nil
	}
	version, found := plan.ObjectVersion, plan.ObjectVersion != ""
	if !found {
		version, found, err = worker.inventory.FindExactVersion(ctx, plan.ObjectKey,
			plan.Generation, plan.CiphertextBytes)
		if err != nil {
			return Result{}, ErrUnavailable
		}
	}
	return worker.finalize(ctx, runner, plan, version, found)
}

func (worker *Worker) prepare(ctx context.Context, runner JobRunner, tenantID,
	uploadID domain.ID, generation uint64,
) (Plan, error) {
	var plan Plan
	err := runner.InRepeatableReadJobTenant(ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			plan, bindErr = store.Prepare(scoped, uploadID, generation,
				worker.clock.Now())
			return bindErr
		})
	return plan, mapError(err)
}

func (worker *Worker) finalize(ctx context.Context, runner JobRunner, plan Plan,
	version string, found bool,
) (Result, error) {
	deletionID, releaseID, err := worker.identities(found)
	if err != nil {
		return Result{}, err
	}
	var finalized Finalization
	err = runner.InRepeatableReadJobTenant(ctx, plan.TenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			finalized, bindErr = store.PersistFinalization(scoped, plan, version, deletionID,
				worker.clock.Now())
			if bindErr != nil {
				return bindErr
			}
			if finalized.ReleaseReservation {
				bindErr = worker.billing.Release(scoped, scope,
					finalized.BillingReservationID, releaseID)
			}
			if bindErr == nil && found {
				payload, _ := json.Marshal(struct {
					DeletionID string `json:"deletion_id"`
				}{DeletionID: deletionID.String()})
				bindErr = worker.jobs.EnqueueDeletion(scoped, scope, deletionID,
					deletionID, payload)
			}
			return bindErr
		})
	return Result{UploadID: plan.UploadID, DeletionID: deletionID,
		State: finalized.State, ObjectFound: found}, mapError(err)
}

func (worker *Worker) identities(found bool) (domain.ID, domain.ID, error) {
	deletionID := domain.ID{}
	if found {
		value, err := worker.ids.NewUUIDv7()
		if err != nil || !value.UUIDv7() {
			return domain.ID{}, domain.ID{}, ErrUnavailable
		}
		deletionID = value
	}
	releaseID, err := worker.ids.NewUUIDv7()
	if err != nil || !releaseID.UUIDv7() || releaseID == deletionID {
		return domain.ID{}, domain.ID{}, ErrUnavailable
	}
	return deletionID, releaseID, nil
}
