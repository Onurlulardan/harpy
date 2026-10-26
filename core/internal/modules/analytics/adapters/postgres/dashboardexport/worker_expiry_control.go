package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) ControlExpiry(
	ctx context.Context, tenantID, exportID, expiryJobID uuid.UUID,
	observedAt time.Time,
) (exportapp.ExpiryControl, error) {
	if tenantID == uuid.Nil || exportID.Version() != 7 ||
		expiryJobID.Version() != 7 || observedAt.IsZero() {
		return exportapp.ExpiryControl{}, exportapp.ErrInvalid
	}
	result := exportapp.ExpiryControl{}
	err := storage.inTenant(ctx, tenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		schema, err := bound.currentSchema(boundContext)
		if err != nil {
			return err
		}
		var state string
		if err = bound.queryRow(boundContext, workerExpiryControlSQL,
			schema, tenantID, exportID, expiryJobID, observedAt,
		).Scan(&state, &result.CleanupRequired); err != nil {
			return err
		}
		result.State = exportapp.State(state)
		if !result.State.Valid() || result.CleanupRequired &&
			result.State != exportapp.StateReady &&
			result.State != exportapp.StateDeleting {
			return exportapp.ErrUnavailable
		}
		return nil
	})
	if err != nil {
		return exportapp.ExpiryControl{}, mapDispatchError(err)
	}
	return result, nil
}
