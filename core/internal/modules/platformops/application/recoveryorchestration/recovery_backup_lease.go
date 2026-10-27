package recoveryorchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const databaseBackupHeartbeat = 30 * time.Second

func (executor *databaseBackupExecutor) withLease(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
	run func(context.Context) error,
) error {
	work, cancel := context.WithCancel(ctx)
	defer cancel()
	heartbeatFailure := make(chan error, 1)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		timer := time.NewTicker(databaseBackupHeartbeat)
		defer timer.Stop()
		for {
			select {
			case <-work.Done():
				return
			case <-timer.C:
				if err := executor.execution.UpdateLease(work, identity); err != nil {
					heartbeatFailure <- err
					cancel()
					return
				}
			}
		}
	}()
	runErr := run(work)
	cancel()
	<-stopped
	select {
	case err := <-heartbeatFailure:
		return fmt.Errorf("database backup lease lost: %w", err)
	default:
	}
	if runErr != nil {
		return runErr
	}
	return nil
}
