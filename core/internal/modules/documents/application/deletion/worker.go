package deletion

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const Operation = "documents.blob.delete"

type Worker struct {
	binder  StoreBinder
	objects ObjectStore
	billing Billing
	ids     foundation.IDSource
	clock   foundation.Clock
}

func NewWorker(binder StoreBinder, objects ObjectStore, billing Billing,
	ids foundation.IDSource, clock foundation.Clock,
) (*Worker, error) {
	for _, dependency := range []any{binder, objects, billing, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents deletion dependencies are invalid")
		}
	}
	return &Worker{binder: binder, objects: objects, billing: billing,
		ids: ids, clock: clock}, nil
}

func (worker *Worker) Run(ctx context.Context, runner JobRunner,
	tenantID, deletionID, leaseOwnerID domain.ID,
) (Result, error) {
	if worker == nil || foundation.NilPort(runner) || !tenantID.Valid() ||
		!deletionID.UUIDv7() ||
		!leaseOwnerID.UUIDv7() || !platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	var task Task
	err := runner.InRepeatableReadJobTenant(ctx, tenantID, func(scoped context.Context,
		scope platformtx.JobTenantScope,
	) error {
		store, bindErr := worker.binder.BindJob(scoped, scope)
		if bindErr != nil || store == nil {
			return ErrUnavailable
		}
		task, bindErr = store.Lease(scoped, deletionID, leaseOwnerID,
			worker.clock.Now(), time.Minute)
		return bindErr
	})
	if err != nil {
		return Result{}, mapError(err)
	}
	receipt, err := worker.objects.DeleteExact(ctx, task)
	if err != nil {
		return Result{}, mapError(err)
	}
	return worker.confirm(ctx, runner, tenantID, task, receipt)
}
