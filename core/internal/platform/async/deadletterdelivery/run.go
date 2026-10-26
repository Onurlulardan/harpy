package deadletterdelivery

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (runner *Runner) run(ctx context.Context) error {
	reconnectCap := runner.settings.ReconnectMinimum
	for {
		if ctx.Err() != nil {
			return nil
		}
		if !runner.waitForAdmission(ctx) {
			return nil
		}
		session, err := runner.connector.Open(ctx)
		if err == nil && session == nil {
			return ErrSessionUnavailable
		}
		if err != nil {
			if err := runner.waitReconnect(ctx, reconnectCap); err != nil {
				return nil
			}
			reconnectCap = nextReconnectCap(reconnectCap, runner.settings.ReconnectMaximum)
			continue
		}
		reconnectCap = runner.settings.ReconnectMinimum
		err, paused := runner.runAdmittedSession(ctx, session)
		if ctx.Err() != nil {
			return nil
		}
		if paused {
			reconnectCap = runner.settings.ReconnectMinimum
			continue
		}
		if err == nil {
			return fmt.Errorf("%w: session stopped without an error", ErrSessionUnavailable)
		}
		if errors.Is(err, ErrDependencies) {
			return err
		}
		if err := runner.waitReconnect(ctx, reconnectCap); err != nil {
			return nil
		}
		reconnectCap = nextReconnectCap(reconnectCap, runner.settings.ReconnectMaximum)
	}
}

func (runner *Runner) waitReconnect(ctx context.Context, cap time.Duration) error {
	delay := reconnectDelay(cap, runner.settings, runner.jitter)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
