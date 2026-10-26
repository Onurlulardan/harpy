package recoveryorchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const databaseRestoreHeartbeat = 30 * time.Second

func (executor *databaseRestoreExecutor) withRestoreLease(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
	run func(context.Context) error,
) error {
	work, cancel := context.WithCancel(ctx)
	defer cancel()
	failure := make(chan error, 1)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(databaseRestoreHeartbeat)
		defer ticker.Stop()
		for {
			select {
			case <-work.Done():
				return
			case <-ticker.C:
				if err := executor.execution.UpdateLease(work, identity); err != nil {
					failure <- err
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
	case err := <-failure:
		return resumableRestore(fmt.Errorf("database restore lease lost: %w", err))
	default:
		return runErr
	}
}
