package jobexecution

import (
	"context"
	"errors"
	"fmt"
	"time"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func (runner *Runner) processClaim(
	forceContext context.Context,
	claim asyncpostgres.JobClaim,
) error {
	processor, err := runner.registry.Resolve(claim)
	if err != nil {
		return fmt.Errorf("resolve claimed job processor: %w", err)
	}
	processContext, cancel := context.WithCancel(forceContext)
	defer cancel()
	done := make(chan error, 1)
	go runProcessor(done, processor, processContext, claim)
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case processErr := <-done:
			if errors.Is(processErr, asyncpostgres.ErrJobAuthorityStale) {
				runner.parkStaleClaim(forceContext, claim)
			}
			return nil
		case <-ticker.C:
			lease, heartbeatErr := runner.store.HeartbeatJob(forceContext, claim)
			if heartbeatErr == nil {
				if lease.CancelRequested() {
					cancel()
				}
				continue
			}
			cancel()
			if errors.Is(heartbeatErr, asyncpostgres.ErrJobLeaseLost) {
				return waitAfterLeaseLoss(done)
			}
			if errors.Is(heartbeatErr, asyncpostgres.ErrJobAuthorityStale) {
				return runner.waitAndParkStaleClaim(forceContext, claim, done)
			}
			return waitAfterLeaseLoss(done)
		case <-forceContext.Done():
			cancel()
			return waitAfterLeaseLoss(done)
		}
	}
}

func (runner *Runner) waitAndParkStaleClaim(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	done <-chan error,
) error {
	<-done
	runner.parkStaleClaim(ctx, claim)
	return nil
}

func (runner *Runner) parkStaleClaim(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) {
	// A failed direct park remains discoverable by the next reconciliation scan.
	_, _ = runner.store.ParkStaleJob(ctx, claim)
}

func runProcessor(
	done chan<- error,
	processor Processor,
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) {
	defer func() {
		if recover() != nil {
			done <- errors.New("job processor panicked")
		}
	}()
	done <- processor.Process(ctx, claim)
}

func waitAfterLeaseLoss(done <-chan error) error {
	<-done
	return nil
}
