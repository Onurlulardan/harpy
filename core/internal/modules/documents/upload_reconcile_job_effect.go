package documents

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
)

type UploadReconcileEffect interface {
	Execute(context.Context, foundation.ID, foundation.ID, uint64) (uploadreconcileapp.Result, error)
}

type uploadReconcileWorkerEffect struct {
	worker *uploadreconcileapp.Worker
	runner uploadreconcileapp.JobRunner
}

func NewUploadReconcileWorkerEffect(worker *uploadreconcileapp.Worker,
	runner uploadreconcileapp.JobRunner,
) (UploadReconcileEffect, error) {
	if worker == nil || nilArchivePort(runner) {
		return nil, fmt.Errorf("Documents upload reconcile effect dependencies are invalid")
	}
	return &uploadReconcileWorkerEffect{worker: worker, runner: runner}, nil
}

func (effect *uploadReconcileWorkerEffect) Execute(ctx context.Context, tenantID,
	uploadID foundation.ID, generation uint64,
) (uploadreconcileapp.Result, error) {
	if effect == nil || effect.worker == nil || nilArchivePort(effect.runner) {
		return uploadreconcileapp.Result{}, uploadreconcileapp.ErrUnavailable
	}
	return effect.worker.Run(ctx, effect.runner, tenantID, uploadID, generation)
}
