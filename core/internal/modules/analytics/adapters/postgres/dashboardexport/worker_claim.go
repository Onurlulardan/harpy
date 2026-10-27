package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) ClaimDelete(
	ctx context.Context, tenantID, exportID, jobID, leaseOwner uuid.UUID,
	observedAt time.Time,
) (exportapp.DeleteClaim, error) {
	if tenantID == uuid.Nil || exportID.Version() != 7 || jobID.Version() != 7 ||
		leaseOwner == uuid.Nil || observedAt.IsZero() {
		return exportapp.DeleteClaim{}, exportapp.ErrInvalid
	}
	var result exportapp.DeleteClaim
	err := storage.inTenant(ctx, tenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		stored := deleteClaimRow{content: contentRow{tenantID: tenantID}}
		if err := bound.queryRow(boundContext, workerDeleteClaimSQL, exportID, jobID,
			leaseOwner, observedAt).Scan(stored.destinations()...); err != nil {
			return err
		}
		decoded, err := stored.decode()
		if err == nil {
			result = decoded
		}
		return err
	})
	if err != nil {
		return exportapp.DeleteClaim{}, mapError(err, nil)
	}
	return result, nil
}
