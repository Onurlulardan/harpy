package dashboardexport

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (worker *Worker) ProcessDelete(
	ctx context.Context, tenantID, exportID, jobID, leaseOwner uuid.UUID,
) (RenderResult, error) {
	if worker == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		!validUUIDv7(exportID) ||
		!validUUIDv7(jobID) || leaseOwner == uuid.Nil {
		return RenderResult{}, ErrInvalid
	}
	claim, err := worker.store.ClaimDelete(
		ctx, tenantID, exportID, jobID, leaseOwner, worker.clock.Now())
	if err != nil {
		return RenderResult{}, err
	}
	if claim.Disposition == ClaimTerminal {
		return RenderResult{State: claim.State}, nil
	}
	// The request path already authorized and pinned the delete intent. Cleanup
	// must continue after permission revocation so retention cannot strand an
	// encrypted artifact indefinitely.
	if claim.Draft == nil {
		receipt := DeleteReceipt{VerifiedAbsentAt: worker.clock.Now()}
		if err = worker.store.CompleteDeleted(ctx, claim, receipt, worker.clock.Now()); err != nil {
			return RenderResult{}, err
		}
		return RenderResult{State: StateDeleted}, nil
	}
	if claim.ObjectVersion == "" {
		reconciled, reconcileErr := worker.artifacts.Reconcile(ctx, *claim.Draft)
		if reconcileErr != nil || reconciled.State == ReconciliationUnknown {
			return RenderResult{State: StateDeleting}, ErrOutcomeUnknown
		}
		if reconciled.State == ReconciliationMissing {
			receipt := DeleteReceipt{ProviderRequestID: reconciled.Receipt.ProviderRequestID,
				AbsenceSHA256: reconciled.EvidenceSHA256, VerifiedAbsentAt: reconciled.VerifiedAt}
			if err = worker.store.CompleteDeleted(ctx, claim, receipt, worker.clock.Now()); err != nil {
				return RenderResult{}, err
			}
			return RenderResult{State: StateDeleted}, nil
		}
		claim.ObjectVersion = reconciled.Receipt.ObjectVersion
	}
	receipt, err := worker.artifacts.Delete(ctx, *claim.Draft, claim.ObjectVersion)
	if errors.Is(err, ErrOutcomeUnknown) {
		reconciled, reconcileErr := worker.artifacts.Reconcile(ctx, *claim.Draft)
		if reconcileErr != nil || reconciled.State == ReconciliationUnknown {
			return RenderResult{}, ErrOutcomeUnknown
		}
		if reconciled.State == ReconciliationSealed {
			return RenderResult{State: StateDeleting}, ErrOutcomeUnknown
		}
		receipt = DeleteReceipt{ProviderRequestID: reconciled.Receipt.ProviderRequestID,
			AbsenceSHA256: reconciled.EvidenceSHA256, VerifiedAbsentAt: reconciled.VerifiedAt}
	} else if err != nil {
		return RenderResult{}, err
	}
	if err = worker.store.CompleteDeleted(ctx, claim, receipt, worker.clock.Now()); err != nil {
		return RenderResult{}, err
	}
	return RenderResult{State: StateDeleted}, nil
}
