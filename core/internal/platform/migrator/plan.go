package migrator

import (
	"context"
	"errors"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func (runner *Runner) Plan(ctx context.Context) (_ PlanResult, resultErr error) {
	if ctx == nil {
		return PlanResult{}, fmt.Errorf("plan context is nil")
	}
	lock, err := runner.acquirePlatformLock(ctx)
	if err != nil {
		return PlanResult{}, err
	}
	defer func() {
		resultErr = errors.Join(resultErr, runner.releasePlatformLock(ctx, lock))
	}()
	history, _, err := readHistory(ctx, lock.connection)
	if err != nil {
		return PlanResult{}, err
	}
	current, pending, err := classifyHistory(history, runner.selected, runner.historyAliases)
	if err != nil {
		return PlanResult{}, err
	}
	tenantPending, err := runner.tenantPlan(ctx, lock.connection)
	if err != nil {
		return PlanResult{}, err
	}
	pending, err = mergeMigrationSelections(pending, tenantPending)
	if err != nil {
		return PlanResult{}, err
	}
	return PlanResult{
		CurrentVersion: current,
		Pending:        append([]postgresv1.SelectedMigration(nil), pending...),
	}, nil
}
