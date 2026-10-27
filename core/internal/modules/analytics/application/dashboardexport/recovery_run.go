package dashboardexport

import (
	"context"
	"fmt"
	"time"
)

const recoveryPollInterval = 250 * time.Millisecond

func (runner *RecoveryRunner) run(scan, force context.Context) error {
	for {
		if scan.Err() != nil {
			return nil
		}
		if !runner.admission.Accepting() {
			if !waitRecoveryPoll(scan) {
				return nil
			}
			continue
		}
		expiredWorked, err := runner.recoverExpired(scan, force)
		if err != nil {
			return fmt.Errorf("finalize expired Analytics dashboard export: %w", err)
		}
		unknownWorked := false
		if scan.Err() == nil {
			unknownWorked, err = runner.recoverUnknown(scan, force)
			if err != nil {
				return fmt.Errorf("recover unknown Analytics dashboard export: %w", err)
			}
		}
		maintenanceWorked := false
		if scan.Err() == nil {
			maintenanceWorked, err = runner.recoverMaintenanceUnknown(scan, force)
			if err != nil {
				return fmt.Errorf("recover unknown Analytics dashboard export maintenance: %w", err)
			}
		}
		if !expiredWorked && !unknownWorked && !maintenanceWorked &&
			!waitRecoveryPoll(scan) {
			return nil
		}
	}
}

func waitRecoveryPoll(ctx context.Context) bool {
	timer := time.NewTimer(recoveryPollInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (runner *RecoveryRunner) reapResolved(ctx context.Context) error {
	if err := runner.spools.Reap(ctx, runner.store, runner.clock.Now()); err != nil {
		return fmt.Errorf("reap resolved Analytics dashboard export spool: %w", err)
	}
	return nil
}
