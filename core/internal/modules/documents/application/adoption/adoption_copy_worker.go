package adoption

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type AdoptionCopyWorker struct {
	binder JobStoreBinder
	copier Copier
	clock  Clock
}

func NewAdoptionCopyWorker(binder JobStoreBinder, copier Copier,
	clock Clock,
) (*AdoptionCopyWorker, error) {
	for _, dependency := range []any{binder, copier, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents adoption copy dependencies are invalid")
		}
	}
	return &AdoptionCopyWorker{binder: binder, copier: copier, clock: clock}, nil
}

func (worker *AdoptionCopyWorker) prepare(ctx context.Context, runner JobRunner,
	authority Authority, sourceVersionID domain.ID,
) (Checkpoint, error) {
	var plan Checkpoint
	err := runner.InRepeatableReadJobTenant(ctx, authority.TenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			plan, bindErr = store.PrepareCopy(scoped, authority, sourceVersionID,
				worker.clock.Now().UTC())
			return bindErr
		})
	if err != nil {
		return Checkpoint{}, mapError(err)
	}
	if !validCheckpoint(plan, authority) || plan.State != CheckpointCopying {
		return Checkpoint{}, ErrUnavailable
	}
	return plan, nil
}

func (worker *AdoptionCopyWorker) finalize(ctx context.Context, runner JobRunner,
	authority Authority, plan Checkpoint, receipt CopyReceipt,
) (Checkpoint, error) {
	var result Checkpoint
	err := runner.InRepeatableReadJobTenant(ctx, authority.TenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			result, bindErr = store.VerifyCopy(scoped, authority, plan, receipt)
			return bindErr
		})
	if err != nil {
		return Checkpoint{}, mapError(err)
	}
	if !validCheckpoint(result, authority) || result.State != CheckpointVerified ||
		result.AppliedLSN != receipt.AppliedLSN {
		return Checkpoint{}, ErrUnavailable
	}
	return result, nil
}
