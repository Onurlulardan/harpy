package processhost

import "context"

func (host *Host) startPoller(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})
	host.pollerCancel = cancel
	host.pollerDone = done
	go func() {
		defer close(done)
		host.dependencies.Poller.Run(ctx)
	}()
}

func (host *Host) waitPoller(ctx context.Context) error {
	if host.pollerDone == nil {
		return nil
	}
	select {
	case <-host.pollerDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
