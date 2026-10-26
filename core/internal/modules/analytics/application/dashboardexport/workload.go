package dashboardexport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type AsyncWorkload interface {
	Start(context.Context) error
	BeginDrain()
	Wait(context.Context) error
	Force() error
	Failures() <-chan error
}

type Workload struct {
	render, recovery AsyncWorkload
	mutex            sync.Mutex
	started          bool
	renderStarted    bool
	recoveryStarted  bool
	failures         chan error
	failuresOnce     sync.Once
}

func NewWorkload(render, recovery AsyncWorkload) (*Workload, error) {
	if nilPort(render) || nilPort(recovery) {
		return nil, fmt.Errorf("Analytics dashboard export workload is unavailable")
	}
	return &Workload{render: render, recovery: recovery,
		failures: make(chan error, 1)}, nil
}

func (workload *Workload) Start(ctx context.Context) error {
	if ctx == nil || ctx.Err() != nil {
		return ErrInvalid
	}
	workload.mutex.Lock()
	if workload.started {
		workload.mutex.Unlock()
		return fmt.Errorf("Analytics dashboard export workload is already started")
	}
	workload.started = true
	workload.mutex.Unlock()
	if err := workload.recovery.Start(ctx); err != nil {
		workload.closeFailures()
		return fmt.Errorf("start Analytics export recovery: %w", err)
	}
	workload.mutex.Lock()
	workload.recoveryStarted = true
	workload.mutex.Unlock()
	if err := workload.render.Start(ctx); err != nil {
		forceErr := workload.recovery.Force()
		waitContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		waitErr := workload.recovery.Wait(waitContext)
		cancel()
		workload.mutex.Lock()
		workload.recoveryStarted = false
		workload.mutex.Unlock()
		workload.closeFailures()
		return errors.Join(fmt.Errorf("start Analytics export render dispatcher: %w", err),
			forceErr, waitErr)
	}
	workload.mutex.Lock()
	workload.renderStarted = true
	workload.mutex.Unlock()
	go workload.monitorFailures()
	return nil
}

func (workload *Workload) monitorFailures() {
	var wait sync.WaitGroup
	for _, active := range []AsyncWorkload{workload.recovery, workload.render} {
		wait.Add(1)
		go func(item AsyncWorkload) {
			defer wait.Done()
			if err, open := <-item.Failures(); open && err != nil {
				select {
				case workload.failures <- err:
				default:
				}
			}
		}(active)
	}
	wait.Wait()
	workload.closeFailures()
}

func (workload *Workload) Failures() <-chan error { return workload.failures }

func (workload *Workload) BeginDrain() {
	render, recovery := workload.activeChildren()
	if recovery != nil {
		recovery.BeginDrain()
	}
	if render != nil {
		render.BeginDrain()
	}
}

func (workload *Workload) Force() error {
	var recoveryErr, renderErr error
	render, recovery := workload.activeChildren()
	if recovery != nil {
		recoveryErr = recovery.Force()
	}
	if render != nil {
		renderErr = render.Force()
	}
	return errors.Join(recoveryErr, renderErr)
}

func (workload *Workload) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalid
	}
	errorsByWorkload := make(chan error, 2)
	count := 0
	render, recovery := workload.activeChildren()
	if recovery != nil {
		count++
		go func() { errorsByWorkload <- recovery.Wait(ctx) }()
	}
	if render != nil {
		count++
		go func() { errorsByWorkload <- render.Wait(ctx) }()
	}
	var result error
	for index := 0; index < count; index++ {
		result = errors.Join(result, <-errorsByWorkload)
	}
	return result
}

func (workload *Workload) activeChildren() (render, recovery AsyncWorkload) {
	workload.mutex.Lock()
	defer workload.mutex.Unlock()
	if workload.renderStarted {
		render = workload.render
	}
	if workload.recoveryStarted {
		recovery = workload.recovery
	}
	return render, recovery
}

func (workload *Workload) closeFailures() {
	workload.failuresOnce.Do(func() { close(workload.failures) })
}

var _ AsyncWorkload = (*Workload)(nil)
var _ AsyncWorkload = (*RecoveryRunner)(nil)
var _ AsyncWorkload = (*RenderDispatcher)(nil)
