package outboxdelivery

import (
	"context"
	"errors"
	"time"
)

func (runner *Runner) run(ctx context.Context) error {
	reconnectCap := runner.settings.ReconnectMinimum
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		broker, err := runner.connector.Open(ctx)
		if err == nil && broker == nil {
			err = ErrDependencies
		}
		if err != nil {
			if err := runner.waitReconnect(ctx, reconnectCap); err != nil {
				return nil
			}
			reconnectCap = nextReconnectCap(reconnectCap, runner.settings.ReconnectMaximum)
			continue
		}
		runner.setActive(broker)
		reconnectCap = runner.settings.ReconnectMinimum
		err = runner.runSession(ctx, broker)
		runner.setActive(nil)
		_ = broker.Close()
		if ctx.Err() != nil {
			return nil
		}
		if fatalSessionError(err) {
			return err
		}
		if err == nil {
			return errors.New("outbox delivery session stopped unexpectedly")
		}
		if err := runner.waitReconnect(ctx, reconnectCap); err != nil {
			return nil
		}
		reconnectCap = nextReconnectCap(reconnectCap, runner.settings.ReconnectMaximum)
	}
}

func (runner *Runner) waitReconnect(ctx context.Context, cap time.Duration) error {
	delay := reconnectDelay(cap, runner.settings, runner.jitter)
	return waitContext(ctx, delay)
}
