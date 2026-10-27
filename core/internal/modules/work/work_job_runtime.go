package work

import (
	"fmt"
	"time"

	jobspostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres/jobs"
	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type WorkJobRuntime struct {
	binding platformasync.DatabaseJobBinding
	now     func() time.Time
}

func NewWorkJobRuntime() (*WorkJobRuntime, error) {
	return newWorkJobRuntime(time.Now)
}

func newWorkJobRuntime(now func() time.Time) (*WorkJobRuntime, error) {
	if now == nil {
		return nil, fmt.Errorf("Work job clock is nil")
	}
	binding, err := platformasync.NewRepeatableReadDatabaseJobBinding("work",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store, storeErr := jobspostgres.NewBoundStore(database)
			runner, runnerErr := jobapp.NewRuntimeRunner(store)
			if storeErr != nil {
				runnerErr = storeErr
			}
			return &workDatabaseJobCommand{runner: runner, buildErr: runnerErr, now: now}
		})
	if err != nil {
		return nil, fmt.Errorf("build Work database job binding: %w", err)
	}
	return &WorkJobRuntime{binding: binding, now: now}, nil
}

func (runtime *WorkJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}
