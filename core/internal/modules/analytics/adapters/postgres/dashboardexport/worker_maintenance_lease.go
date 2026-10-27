package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) CheckpointMaintenanceUnknown(
	ctx context.Context, claim exportapp.MaintenanceUnknownClaim, observedAt time.Time,
) (exportapp.UnknownCleanupLease, error) {
	if !validMaintenanceClaim(claim) || observedAt.IsZero() {
		return exportapp.UnknownCleanupLease{}, exportapp.ErrInvalid
	}
	var leaseUntil time.Time
	err := storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := maintenanceSchema(boundContext, bound, claim)
			if err != nil {
				return err
			}
			arguments := append(maintenanceClaimArguments(schema, claim), observedAt)
			return bound.queryRow(boundContext, workerMaintenanceUnknownCheckpointSQL,
				arguments...).Scan(&leaseUntil)
		})
	if err != nil {
		return exportapp.UnknownCleanupLease{}, mapDispatchError(err)
	}
	leaseUntil = leaseUntil.UTC()
	if !leaseUntil.After(observedAt) {
		return exportapp.UnknownCleanupLease{}, exportapp.ErrUnavailable
	}
	return exportapp.UnknownCleanupLease{LeaseUntil: leaseUntil}, nil
}

func (storage *WorkerStore) DeferMaintenanceUnknown(
	ctx context.Context, claim exportapp.MaintenanceUnknownClaim, observedAt time.Time,
) (time.Time, error) {
	if !validMaintenanceClaim(claim) || observedAt.IsZero() {
		return time.Time{}, exportapp.ErrInvalid
	}
	var retryAt time.Time
	err := storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, err := maintenanceSchema(boundContext, bound, claim)
			if err != nil {
				return err
			}
			arguments := append(maintenanceClaimArguments(schema, claim), observedAt)
			return bound.queryRow(boundContext, workerMaintenanceUnknownDeferSQL,
				arguments...).Scan(&retryAt)
		})
	if err != nil {
		return time.Time{}, mapDispatchError(err)
	}
	return retryAt.UTC(), nil
}

func maintenanceSchema(ctx context.Context, bound *workerBound,
	claim exportapp.MaintenanceUnknownClaim,
) (string, error) {
	schema, err := bound.currentSchema(ctx)
	if err != nil || schema != claim.Candidate.TenantSchema {
		return "", exportapp.ErrUnavailable
	}
	return schema, nil
}

func maintenanceClaimArguments(schema string,
	claim exportapp.MaintenanceUnknownClaim,
) []any {
	return append(maintenanceCandidateArguments(schema, claim.Candidate),
		claim.CleanupGeneration, claim.CleanupLeaseOwnerID,
		claim.DeleteAttemptGeneration)
}
