package migrator

import (
	"context"
	"errors"
	"fmt"
)

func (runner *Runner) Verify(ctx context.Context) (_ VerifyResult, resultErr error) {
	if ctx == nil {
		return VerifyResult{}, fmt.Errorf("verify context is nil")
	}
	lock, err := runner.acquirePlatformLock(ctx)
	if err != nil {
		return VerifyResult{}, err
	}
	defer func() {
		resultErr = errors.Join(resultErr, runner.releasePlatformLock(ctx, lock))
	}()
	history, exists, err := readHistory(ctx, lock.connection)
	if err != nil {
		return VerifyResult{}, err
	}
	current, pending, err := classifyHistory(history, runner.selected, runner.historyAliases)
	if err != nil {
		return VerifyResult{}, err
	}
	if !exists || len(pending) != 0 {
		return VerifyResult{}, errors.Join(
			ErrMigrationsPending,
			fmt.Errorf("%d platform migrations are pending", len(pending)),
		)
	}
	tenantPending, err := runner.tenantPlan(ctx, lock.connection)
	if err != nil {
		return VerifyResult{}, err
	}
	if len(tenantPending) != 0 {
		return VerifyResult{}, errors.Join(
			ErrMigrationsPending,
			fmt.Errorf("%d tenant migrations are pending", len(tenantPending)),
		)
	}
	if err := requireLifecycleReady(ctx, lock.connection); err != nil {
		return VerifyResult{}, err
	}
	if err := runner.verifyTenantMigrations(ctx, lock.connection); err != nil {
		return VerifyResult{}, err
	}
	contract, err := runner.verifyStoredEvidence(ctx, lock.connection)
	if err != nil {
		return VerifyResult{}, err
	}
	return VerifyResult{
		DatabaseVersion:          current,
		DatabaseGeneration:       contract.DatabaseGeneration,
		MinimumRuntimeGeneration: contract.MinimumRuntimeGeneration,
		VerifiedRunID:            contract.VerifiedRunID,
	}, nil
}
