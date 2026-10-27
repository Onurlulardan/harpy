package executor

import (
	"context"
	"sync"
)

func (fixture *graphFakeFixture) execute(ctx context.Context) (Result, error) {
	run := fixture.startPrograms(ctx)
	barrierErr, programErr := runGraphLifecycle(ctx, fixture.runtime.barrier, run)
	if barrierErr != nil {
		return Result{}, barrierErr
	}
	if programErr != nil {
		return Result{}, programErr
	}
	return fixture.runtime.result()
}

func (fixture *graphFakeFixture) startPrograms(ctx context.Context) *graphProgramRun {
	run := &graphProgramRun{done: make(chan struct{})}
	runContext, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		var wait sync.WaitGroup
		errors := make(chan error, len(fixture.runtime.nodes)+1)
		start := func(callback func() error) {
			wait.Add(1)
			go func() {
				defer wait.Done()
				err := callback()
				if err != nil {
					cancel()
				}
				errors <- err
			}()
		}
		start(func() error {
			return fixture.runtime.runRootProgram(runContext, fixture.drivers[0])
		})
		for index := range fixture.runtime.nodes {
			current := index
			start(func() error {
				return fixture.runtime.runNodeProgram(
					runContext, fixture.drivers[current+1], current,
				)
			})
		}
		wait.Wait()
		close(errors)
		for err := range errors {
			if run.err == nil && err != nil {
				run.err = err
			}
		}
		close(run.done)
	}()
	return run
}

func graphCallIndex(calls []string, kind string, value statement) int {
	target := kind + "|" + value.sql
	if kind == "row" {
		target = kind + "|" + value.explainSQL
	}
	for index, call := range calls {
		if call == target {
			return index
		}
	}
	return -1
}
