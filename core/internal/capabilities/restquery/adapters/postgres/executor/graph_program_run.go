package executor

import (
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
)

type graphProgramRun struct {
	done       chan struct{}
	once       sync.Once
	err        error
	panicValue any
}

func startGraphPrograms(
	registry execution.Registry,
	stage execution.ProgramStage,
) *graphProgramRun {
	run := &graphProgramRun{done: make(chan struct{})}
	go func() {
		run.once.Do(func() {
			defer func() {
				run.panicValue = recover()
				close(run.done)
			}()
			run.err = registry.RunStage(stage)
		})
	}()
	return run
}

func (run *graphProgramRun) outcome() error {
	if run == nil || run.done == nil {
		return ErrContract
	}
	<-run.done
	if run.panicValue != nil {
		panic(run.panicValue)
	}
	return run.err
}
