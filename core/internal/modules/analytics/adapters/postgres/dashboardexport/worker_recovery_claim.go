package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) ClaimUnknown(
	ctx context.Context, candidate exportapp.UnknownCandidate,
	cleanupOwner uuid.UUID, observedAt time.Time,
) (exportapp.UnknownCleanupClaim, error) {
	if !validUnknownCandidate(candidate) || cleanupOwner.Version() != 7 ||
		observedAt.IsZero() {
		return exportapp.UnknownCleanupClaim{}, exportapp.ErrInvalid
	}
	result := exportapp.UnknownCleanupClaim{}
	err := storage.inTenant(ctx, candidate.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		schema, err := bound.currentSchema(boundContext)
		if err != nil || schema != candidate.TenantSchema {
			return exportapp.ErrUnavailable
		}
		stored := recoveryClaimRow{}
		err = bound.queryRow(boundContext, workerUnknownCleanupClaimSQL,
			schema, candidate.TenantID, candidate.ExportID, candidate.JobID,
			candidate.AttemptID, candidate.AttemptGeneration,
			candidate.LeaseOwnerID, candidate.EffectID,
			candidate.EffectRequestSHA256[:], cleanupOwner, observedAt,
		).Scan(stored.targets()...)
		if err != nil {
			return err
		}
		result, err = stored.decode(candidate, cleanupOwner)
		return err
	})
	if err != nil {
		return exportapp.UnknownCleanupClaim{}, mapDispatchError(err)
	}
	return result, nil
}
