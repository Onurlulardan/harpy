package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) ClaimRenderCandidate(ctx context.Context,
	candidate exportapp.RenderCandidate, reservation exportapp.SpoolReservationIdentity,
) (exportapp.RenderDispatchAttempt, error) {
	if !validWorkerCandidate(candidate) || reservation.ExportID != candidate.ExportID ||
		reservation.AttemptGeneration != candidate.AttemptGeneration ||
		reservation.LeaseOwnerID.Version() != 7 {
		return exportapp.RenderDispatchAttempt{}, exportapp.ErrInvalid
	}
	var result exportapp.RenderDispatchAttempt
	err := storage.inTenant(ctx, candidate.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		if bound.currentSchema == nil {
			return exportapp.ErrUnavailable
		}
		schema, err := bound.currentSchema(boundContext)
		if err != nil || schema != candidate.TenantSchema {
			return exportapp.ErrDispatchCandidateChanged
		}
		var stored dispatchClaimRow
		err = bound.queryRow(boundContext, workerClaimSQL, schema,
			candidate.TenantID, candidate.ExportID, candidate.JobID,
			candidate.AttemptGeneration, reservation.LeaseOwnerID,
		).Scan(stored.targets()...)
		if err != nil {
			return err
		}
		decoded, err := stored.decode(candidate, reservation)
		if err == nil {
			result = decoded
		}
		return err
	})
	if err != nil {
		return exportapp.RenderDispatchAttempt{}, mapDispatchError(err)
	}
	return result, nil
}
