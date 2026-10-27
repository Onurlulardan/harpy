package dashboardexport

import (
	"context"
	"sync"
)

type renderDispatchGroup struct {
	slots    chan struct{}
	released chan struct{}
	failures chan error
	wait     sync.WaitGroup
	force    context.CancelFunc
}

func newRenderDispatchGroup(force context.CancelFunc) *renderDispatchGroup {
	return &renderDispatchGroup{slots: make(chan struct{}, renderDispatcherConcurrency),
		released: make(chan struct{}, 1), failures: make(chan error, 1), force: force}
}

func (group *renderDispatchGroup) available() bool {
	return len(group.slots) < cap(group.slots)
}

func (group *renderDispatchGroup) launch(
	dispatcher *RenderDispatcher, ctx context.Context, dispatch renderDispatch,
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
		if err := dispatcher.execute(ctx, dispatch); err != nil {
			select {
			case group.failures <- err:
				group.force()
			default:
			}
		}
	}()
}
