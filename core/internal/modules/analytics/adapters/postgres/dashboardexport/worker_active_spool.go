package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) ActiveSpoolReservation(ctx context.Context,
	reservation exportapp.SpoolReservationIdentity,
) (bool, error) {
	if storage == nil || storage.source == nil || ctx == nil || ctx.Err() != nil ||
		reservation.ExportID.Version() != 7 || reservation.LeaseOwnerID.Version() != 7 ||
		reservation.AttemptGeneration < 1 ||
		reservation.AttemptGeneration > exportapp.MaximumAttempts {
		return false, exportapp.ErrInvalid
	}
	tenantID, found, err := storage.source.AnalyticsDashboardExportSpoolTenant(
		ctx, reservation.ExportID)
	if err != nil {
		return false, exportapp.ErrUnavailable
	}
	if !found {
		return false, nil
	}
	var active bool
	err = storage.inTenant(ctx, tenantID,
		func(boundContext context.Context, bound *workerBound) error {
			if bound.currentSchema == nil {
				return exportapp.ErrUnavailable
			}
			schema, err := bound.currentSchema(boundContext)
			if err != nil || !workerTenantSchemaPattern.MatchString(schema) {
				return exportapp.ErrUnavailable
			}
			return bound.queryRow(boundContext, workerActiveSpoolSQL, schema,
				tenantID, reservation.ExportID, reservation.AttemptGeneration,
				reservation.LeaseOwnerID).Scan(&active)
		})
	return active, mapDispatchError(err)
}
