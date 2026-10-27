package adoption

import (
	"context"

	"github.com/google/uuid"
)

func (worker *Worker) Start(ctx context.Context, request StartRequest) (Run, bool, error) {
	if worker == nil || ctx == nil || ctx.Err() != nil || request.OperationID == uuid.Nil ||
		request.RunID == uuid.Nil || request.RunID.Version() != 7 {
		return Run{}, false, ErrInvalid
	}
	manifest, err := worker.source.Inspect(ctx, request.Locator)
	if err != nil || !manifest.valid() {
		return Run{}, false, ErrSourceDrift
	}
	now := worker.clock.Now().UTC()
	if !validObservedAt(now) {
		return Run{}, false, ErrUnavailable
	}
	run, replayed, err := worker.target.AdoptionStart(ctx, StartCommand{OperationID: request.OperationID,
		RunID: request.RunID, Locator: request.Locator, Manifest: manifest,
		TransformSHA256: TransformSHA256(), ObservedAt: now})
	if err != nil {
		return Run{}, false, err
	}
	if !run.valid() || run.OperationID != request.OperationID || run.RunID != request.RunID ||
		run.ManifestSHA256 != manifest.ManifestSHA256 || run.TransformSHA256 != TransformSHA256() {
		return Run{}, false, ErrUnavailable
	}
	return run, replayed, nil
}
