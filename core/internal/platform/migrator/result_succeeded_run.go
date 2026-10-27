package migrator

import (
	"context"
	"database/sql"
	"fmt"
)

func (runner *Runner) resultForSucceededRun(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
	initial map[int64]historyRow,
) (UpResult, error) {
	history, _, err := readHistory(ctx, connection)
	if err != nil {
		return UpResult{}, err
	}
	current, pending, err := classifyHistory(history, runner.selected, runner.historyAliases)
	if err != nil {
		return UpResult{}, err
	}
	if len(pending) != 0 {
		return UpResult{}, fmt.Errorf("succeeded migration run has pending history")
	}
	tenantPending, err := runner.tenantPlan(ctx, connection)
	if err != nil {
		return UpResult{}, err
	}
	if len(tenantPending) != 0 {
		return UpResult{}, fmt.Errorf(
			"succeeded migration run cannot satisfy pending tenant migrations; start a new run",
		)
	}
	if err := runner.verifyTenantMigrations(ctx, connection); err != nil {
		return UpResult{}, err
	}
	if _, err := runner.verifyStoredEvidence(ctx, connection); err != nil {
		return UpResult{}, err
	}
	return UpResult{
		RunID: runID, DatabaseVersion: current,
		AppliedVersions: appliedVersionDifference(initial, history, runner.selected),
	}, nil
}
