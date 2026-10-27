package documents

import (
	"context"
	"fmt"
	"reflect"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type ArchiveEffect interface {
	Execute(context.Context, foundation.ID, foundation.ID, foundation.ID) (archiveapp.Snapshot, error)
}

type archiveWorkerEffect struct {
	worker *archiveapp.Worker
	runner archiveapp.JobRunner
}

func NewArchiveWorkerEffect(worker *archiveapp.Worker,
	runner archiveapp.JobRunner,
) (ArchiveEffect, error) {
	if worker == nil || nilArchivePort(runner) {
		return nil, fmt.Errorf("Documents archive effect dependencies are invalid")
	}
	return &archiveWorkerEffect{worker: worker, runner: runner}, nil
}

func (effect *archiveWorkerEffect) Execute(ctx context.Context, tenantID,
	archiveID, jobID foundation.ID,
) (archiveapp.Snapshot, error) {
	if effect == nil || effect.worker == nil || nilArchivePort(effect.runner) {
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	return effect.worker.Run(ctx, effect.runner, tenantID, archiveID, jobID)
}

func nilArchivePort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
