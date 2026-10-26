package documents

import (
	"context"
	"fmt"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type DeletionEffect interface {
	Execute(context.Context, foundation.ID, foundation.ID, foundation.ID) (deletionapp.Result, error)
}

type deletionWorkerEffect struct {
	worker *deletionapp.Worker
	runner deletionapp.JobRunner
}

func NewDeletionWorkerEffect(worker *deletionapp.Worker,
	runner deletionapp.JobRunner,
) (DeletionEffect, error) {
	if worker == nil || nilArchivePort(runner) {
		return nil, fmt.Errorf("Documents deletion effect dependencies are invalid")
	}
	return &deletionWorkerEffect{worker: worker, runner: runner}, nil
}

func (effect *deletionWorkerEffect) Execute(ctx context.Context, tenantID,
	deletionID, jobID foundation.ID,
) (deletionapp.Result, error) {
	if effect == nil || effect.worker == nil || nilArchivePort(effect.runner) {
		return deletionapp.Result{}, deletionapp.ErrUnavailable
	}
	return effect.worker.Run(ctx, effect.runner, tenantID, deletionID, jobID)
}
