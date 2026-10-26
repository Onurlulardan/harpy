package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (storage *store) Claim(
	ctx context.Context, profileID, leaseID, runID, expiryID uuid.UUID,
	ttl time.Duration,
) (tenantmigrationexecutor.Claim, bool, error) {
	if err := storage.usable(ctx); err != nil {
		return tenantmigrationexecutor.Claim{}, false, err
	}
	row := storage.pool.QueryRow(
		ctx, claimSQL, profileID, leaseID, runID, expiryID, seconds(ttl),
	)
	var claim tenantmigrationexecutor.Claim
	var previousRunID *uuid.UUID
	err := row.Scan(
		&claim.JobID, &claim.OperationID, &claim.TenantID, &claim.SchemaName,
		&claim.StepPlanSHA256, &claim.Origin, &claim.Attempt,
		&claim.MigrationRunID, &previousRunID, &claim.LeaseID,
		&claim.ProfileInstanceID, &claim.ReleaseID, &claim.EnvironmentID,
		&claim.DeploymentMode, &claim.RuntimeGeneration,
		&claim.DatabaseGeneration, &claim.MigrationManifestSHA256,
		&claim.LeaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return tenantmigrationexecutor.Claim{}, false, nil
	}
	if previousRunID != nil {
		claim.PreviousRunID = *previousRunID
	}
	return claim, err == nil, err
}

func (storage *store) Heartbeat(
	ctx context.Context, claim tenantmigrationexecutor.Claim, ttl time.Duration,
) (time.Time, error) {
	if err := storage.usable(ctx); err != nil {
		return time.Time{}, err
	}
	var until time.Time
	err := storage.pool.QueryRow(ctx, heartbeatSQL, claim.JobID, claim.Attempt,
		claim.ProfileInstanceID, claim.LeaseID, seconds(ttl)).Scan(&until)
	return until, err
}
