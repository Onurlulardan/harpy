package tenantmigrationexecutor

import (
	"context"
	"time"
)

func (executor *Executor) heartbeatClaim(
	ctx context.Context, claim Claim, cancel context.CancelFunc, done chan<- error,
) {
	ticker := time.NewTicker(LeaseHeartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			done <- nil
			return
		case <-ticker.C:
			if _, err := executor.store.Heartbeat(ctx, claim, executor.leaseTTL); err != nil {
				cancel()
				done <- err
				return
			}
		}
	}
}
