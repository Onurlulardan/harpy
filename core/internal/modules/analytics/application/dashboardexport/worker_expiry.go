package dashboardexport

import (
	"context"

	"github.com/google/uuid"
)

func (worker *Worker) ProcessExpiry(
	ctx context.Context, tenantID, exportID, jobID, leaseOwner uuid.UUID,
) (RenderResult, error) {
	if worker == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		!validUUIDv7(exportID) || !validUUIDv7(jobID) || leaseOwner == uuid.Nil {
		return RenderResult{}, ErrInvalid
	}
	control, err := worker.store.ControlExpiry(
		ctx, tenantID, exportID, jobID, worker.clock.Now())
	if err != nil {
		return RenderResult{}, err
	}
	if !control.CleanupRequired {
		return RenderResult{State: control.State}, nil
	}
	return worker.ProcessDelete(ctx, tenantID, exportID, jobID, leaseOwner)
}
