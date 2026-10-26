package documents

import (
	"context"
	"fmt"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type AdoptionCopyEffect interface {
	Execute(context.Context, adoptionapp.Authority, foundation.ID) (adoptionapp.Checkpoint, error)
}

type adoptionCopyWorkerEffect struct {
	worker *adoptionapp.AdoptionCopyWorker
	runner adoptionapp.JobRunner
}

func NewAdoptionCopyWorkerEffect(worker *adoptionapp.AdoptionCopyWorker,
	runner adoptionapp.JobRunner,
) (AdoptionCopyEffect, error) {
	if worker == nil || nilArchivePort(runner) {
		return nil, fmt.Errorf("Documents adoption copy effect dependencies are invalid")
	}
	return &adoptionCopyWorkerEffect{worker: worker, runner: runner}, nil
}

func (effect *adoptionCopyWorkerEffect) Execute(ctx context.Context,
	authority adoptionapp.Authority, sourceVersionID foundation.ID,
) (adoptionapp.Checkpoint, error) {
	if effect == nil || effect.worker == nil || nilArchivePort(effect.runner) {
		return adoptionapp.Checkpoint{}, adoptionapp.ErrUnavailable
	}
	return effect.worker.Copy(ctx, effect.runner, authority, sourceVersionID)
}
