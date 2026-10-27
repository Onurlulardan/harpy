package systemtransfer

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	sourceCleanupIdlePoll      = 5 * time.Second
	sourceCleanupAdmissionPoll = 100 * time.Millisecond
)

func (runner *SourceCleanupRunner) run(scan, force context.Context) error {
	var after uuid.UUID
	for {
		if scan.Err() != nil {
			return nil
		}
		if !waitForSourceCleanupAdmission(scan, runner.admission) {
			return nil
		}
		result, err := ClaimSourceCleanup(scan, runner.store,
			SourceCleanupClaim{AfterTenantID: after,
				LeaseOwnerID: runner.ownerID, ObservedAt: runner.clock.Now().UTC()})
		if err != nil {
			if scan.Err() != nil {
				return nil
			}
			return err
		}
		if result.ScannedTenantID == uuid.Nil {
			after = uuid.Nil
			if !waitForSourceCleanup(scan, sourceCleanupIdlePoll) {
				return nil
			}
			continue
		}
		after = result.ScannedTenantID
		if result.Found {
			if err := runner.execute(force, result.Target); err != nil {
				return err
			}
		}
	}
}

func waitForSourceCleanupAdmission(ctx context.Context, gate SourceCleanupAdmission) bool {
	for {
		if ctx.Err() != nil {
			return false
		}
		if gate.Accepting() {
			return true
		}
		if !waitForSourceCleanup(ctx, sourceCleanupAdmissionPoll) {
			return false
		}
	}
}

func waitForSourceCleanup(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
