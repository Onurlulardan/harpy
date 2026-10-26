package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) recoverFinalizedRelease(
	ctx context.Context,
	connection *sql.Conn,
	currentVersion int64,
	pendingCount int,
) (UpResult, bool, error) {
	exists, err := postgresstore.ReleaseGenerationTableExists(ctx, connection)
	if err != nil || !exists {
		return UpResult{}, false, err
	}
	stored, err := postgresstore.FindReleaseGenerationByRelease(
		ctx, connection, runner.identity.environmentID, runner.identity.releaseID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return UpResult{}, false, nil
	}
	if err != nil {
		return UpResult{}, false, err
	}
	expected, err := runner.currentReleaseGeneration(stored.MigrationRunID)
	if err != nil {
		return UpResult{}, false, err
	}
	if stored != expected {
		return UpResult{}, false, fmt.Errorf("finalized release tuple differs")
	}
	if pendingCount != 0 {
		return UpResult{}, false, fmt.Errorf("finalized release has pending migrations")
	}
	tenantPending, err := runner.tenantPlan(ctx, connection)
	if err != nil {
		return UpResult{}, false, err
	}
	if len(tenantPending) != 0 {
		return UpResult{}, false, fmt.Errorf(
			"finalized release cannot satisfy pending tenant migrations; use tenant onboarding",
		)
	}
	if err := runner.verifyTenantMigrations(ctx, connection); err != nil {
		return UpResult{}, false, err
	}
	contract, err := runner.verifyStoredEvidence(ctx, connection)
	if err != nil {
		return UpResult{}, false, err
	}
	if contract.VerifiedRunID != stored.MigrationRunID {
		return UpResult{}, false, fmt.Errorf("finalized release is not the active contract")
	}
	return UpResult{
		RunID: stored.MigrationRunID, DatabaseVersion: currentVersion,
		AppliedVersions: []int64{},
	}, true, nil
}
