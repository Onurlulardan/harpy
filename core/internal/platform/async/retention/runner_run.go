package retention

import (
	"context"
	"fmt"
	"time"
)

func (runner *Runner) run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		if _, admitted := runner.admission.Admit(); !admitted {
			if err := runner.wait(ctx, backlogYield); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("wait for async retention admission: %w", err)
			}
			continue
		}
		backlog, err := runner.runWake(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("run async retention wake: %w", err)
		}
		delay := idleInterval
		if backlog {
			delay = backlogYield
		}
		if err := runner.wait(ctx, delay); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("wait for async retention wake: %w", err)
		}
	}
}

func (runner *Runner) runWake(ctx context.Context) (bool, error) {
	for cycle := 0; cycle < maximumCyclesPerWake; cycle++ {
		if _, admitted := runner.admission.Admit(); !admitted {
			return false, nil
		}
		startedAt := runner.now()
		result, err := runner.store.RunRetentionCycle(ctx)
		duration := runner.now().Sub(startedAt)
		if err != nil {
			if ctx.Err() != nil {
				return false, nil
			}
			runner.observer.ObserveFailure(ctx, duration)
			return false, err
		}
		outcome := OutcomeWorked
		if !result.LockAcquired() {
			outcome = OutcomeContended
		} else if !result.Worked() {
			outcome = OutcomeIdle
		}
		observation, err := NewCycleObservation(outcome, duration, result.Counts())
		if err != nil {
			runner.observer.ObserveFailure(ctx, duration)
			return false, err
		}
		runner.observer.ObserveCycle(ctx, observation)
		if outcome != OutcomeWorked {
			return false, nil
		}
	}
	runner.observer.ObserveBacklogYield(ctx)
	return true, nil
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
