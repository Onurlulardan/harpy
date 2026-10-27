package eventdelivery

import (
	"context"
	"fmt"
	"sync"
)

func (runner *Runner) run(parent context.Context) error {
	if runner.registry.Empty() {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	failures := make(chan error, 1)
	group := &sync.WaitGroup{}
	for _, registration := range runner.registry.registrations {
		current := registration
		group.Add(1)
		go func() {
			defer group.Done()
			if err := runner.runRegistrationSafely(ctx, current); err != nil && ctx.Err() == nil {
				select {
				case failures <- err:
					cancel()
				default:
				}
			}
		}()
	}
	group.Wait()
	select {
	case failure := <-failures:
		return failure
	default:
		return nil
	}
}

func (runner *Runner) runRegistrationSafely(
	ctx context.Context,
	registration Registration,
) (result error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = fmt.Errorf(
				"consumer %s version %d: %w",
				registration.consumer, registration.version, ErrSessionPanicked,
			)
		}
	}()
	return runner.runRegistration(ctx, registration)
}
