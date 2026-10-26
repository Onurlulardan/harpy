package archive

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Worker struct {
	binder         StoreBinder
	authorizer     RefreshAuthorizer
	jobs           JobControl
	billing        Billing
	writer         ArtifactWriter
	clock          foundation.Clock
	reservationTTL time.Duration
}

func NewWorker(binder StoreBinder, authorizer RefreshAuthorizer, jobs JobControl, billing Billing,
	writer ArtifactWriter, clock foundation.Clock, reservationTTL time.Duration,
) (*Worker, error) {
	for _, dependency := range []any{binder, authorizer, jobs, billing, writer, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents archive worker dependencies are invalid")
		}
	}
	if reservationTTL < 2*time.Hour || reservationTTL > 48*time.Hour {
		return nil, fmt.Errorf("documents archive reservation TTL is invalid")
	}
	return &Worker{binder: binder, authorizer: authorizer, jobs: jobs,
		billing: billing, writer: writer, clock: clock,
		reservationTTL: reservationTTL}, nil
}

func (worker *Worker) Run(ctx context.Context, runner JobRunner, tenantID,
	archiveID, jobID domain.ID,
) (Snapshot, error) {
	if worker == nil || foundation.NilPort(runner) || !tenantID.Valid() ||
		!archiveID.UUIDv7() || !jobID.UUIDv7() || !platformtx.NetworkAllowed(ctx) {
		return Snapshot{}, ErrInvalid
	}
	plan, err := worker.prepare(ctx, runner, tenantID, archiveID, jobID)
	if err != nil {
		return Snapshot{}, err
	}
	refresh := worker.refreshGate(ctx, plan, jobID)
	var receipt ArtifactReceipt
	if plan.ObjectVersion != "" {
		receipt, err = worker.writer.Recover(ctx, plan)
	} else {
		receipt, err = worker.writer.Write(ctx, plan, refresh)
		if err == nil {
			plan, err = worker.recordEffect(ctx, runner, plan, receipt)
		}
	}
	if err != nil {
		return Snapshot{}, mapError(err)
	}
	if err = refresh(len(plan.Entries)); err != nil || !validReceipt(plan, receipt) {
		if err != nil {
			return Snapshot{}, err
		}
		return Snapshot{}, ErrUnavailable
	}
	return worker.finalize(ctx, runner, plan, receipt)
}

func (worker *Worker) prepare(ctx context.Context, runner JobRunner, tenantID,
	archiveID, jobID domain.ID,
) (Plan, error) {
	var plan Plan
	err := runner.InRepeatableReadJobTenant(ctx, tenantID, func(scoped context.Context,
		scope platformtx.JobTenantScope,
	) error {
		store, bindErr := worker.binder.BindJob(scoped, scope)
		if bindErr != nil || store == nil {
			return ErrUnavailable
		}
		now := worker.clock.Now().UTC()
		plan, bindErr = store.Pin(scoped, archiveID, now)
		if bindErr != nil {
			return bindErr
		}
		if plan.JobID != jobID || plan.TenantID != tenantID || len(plan.Entries) == 0 {
			return ErrConflict
		}
		upper, boundErr := worker.writer.UpperBound(plan)
		if boundErr != nil || upper < 1 {
			return ErrLimit
		}
		ciphertext, boundErr := domain.ArchiveCiphertextUpperBound(upper)
		if boundErr != nil {
			return ErrLimit
		}
		if bindErr = worker.billing.Reserve(scoped, scope, plan.BillingReservationID,
			plan.ID, ciphertext, now.Add(worker.reservationTTL)); bindErr != nil {
			return bindErr
		}
		return worker.billing.StartEffect(scoped, scope,
			plan.BillingReservationID, plan.JobID)
	})
	return plan, mapError(err)
}

func (worker *Worker) finalize(ctx context.Context, runner JobRunner, plan Plan,
	receipt ArtifactReceipt,
) (Snapshot, error) {
	var result Snapshot
	err := runner.InRepeatableReadJobTenant(ctx, plan.TenantID, func(scoped context.Context,
		scope platformtx.JobTenantScope,
	) error {
		store, bindErr := worker.binder.BindJob(scoped, scope)
		if bindErr != nil || store == nil {
			return ErrUnavailable
		}
		if bindErr = worker.billing.Commit(scoped, scope, plan.BillingReservationID,
			plan.ID, receipt.CiphertextBytes); bindErr != nil {
			return bindErr
		}
		result, bindErr = store.PersistReady(scoped, plan.ID, receipt)
		return bindErr
	})
	return result, mapError(err)
}

func (worker *Worker) recordEffect(ctx context.Context, runner JobRunner, plan Plan,
	receipt ArtifactReceipt,
) (Plan, error) {
	var recorded Plan
	err := runner.InRepeatableReadJobTenant(ctx, plan.TenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			recorded, bindErr = store.PersistEffect(scoped, plan.ID, receipt)
			return bindErr
		})
	return recorded, mapError(err)
}

func validReceipt(plan Plan, receipt ArtifactReceipt) bool {
	expectedKey, err := domain.ArchiveObjectKey(plan.TenantID, plan.ID)
	return err == nil && receipt.ObjectKey == expectedKey && receipt.ObjectVersion != "" &&
		receipt.ObjectGeneration > 0 && receipt.PlaintextBytes > plan.PlaintextBytes &&
		receipt.CiphertextBytes > receipt.PlaintextBytes &&
		receipt.PlaintextSHA256 != [32]byte{} && receipt.CiphertextSHA256 != [32]byte{} &&
		receipt.KeyID != "" && len(receipt.WrappedDEK) >= 115 && !receipt.WrittenAt.IsZero()
}
