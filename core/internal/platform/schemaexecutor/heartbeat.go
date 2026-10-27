package schemaexecutor

import (
	"context"
	"time"
)

func (executor *Executor) heartbeat(
	ctx context.Context,
	cancelApply context.CancelFunc,
	lease Lease,
	done chan<- error,
) {
	timer := time.NewTimer(executor.heartbeatInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			done <- nil
			return
		case <-timer.C:
			err := executor.persistence.Heartbeat(ctx, HeartbeatRequest{Lease: lease})
			if err != nil {
				cancelApply()
				done <- err
				return
			}
			timer.Reset(executor.heartbeatInterval)
		}
	}
}
