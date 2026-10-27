package jobexecution

import (
	"context"
	"sync"
)

type executionGroup struct {
	slots    chan struct{}
	released chan struct{}
	failures chan error
	wait     *sync.WaitGroup
	force    context.CancelFunc
	forceCtx context.Context
}

func newExecutionGroup(
	forceContext context.Context,
	forceCancel context.CancelFunc,
) *executionGroup {
	return &executionGroup{
		slots: make(chan struct{}, maximumConcurrent), released: make(chan struct{}, 1),
		failures: make(chan error, 1), wait: &sync.WaitGroup{},
		force: forceCancel, forceCtx: forceContext,
	}
}

func (runner *Runner) run(
	drainContext context.Context,
	forceContext context.Context,
	forceCancel context.CancelFunc,
) error {
	group := newExecutionGroup(forceContext, forceCancel)
	err := runner.scanLoop(drainContext, group)
	if err != nil {
		forceCancel()
	}
	group.wait.Wait()
	if err != nil {
		return err
	}
	select {
	case failure := <-group.failures:
		return failure
	default:
		return nil
	}
}

func (group *executionGroup) launch(
	runner *Runner,
	claim jobClaim,
) {
	group.slots <- struct{}{}
	group.wait.Add(1)
	go func() {
		defer group.wait.Done()
		defer func() {
			<-group.slots
			select {
			case group.released <- struct{}{}:
			default:
			}
		}()
		if err := runner.processClaim(group.forceCtx, claim.value); err != nil {
			select {
			case group.failures <- err:
				group.force()
			default:
			}
		}
	}()
}
