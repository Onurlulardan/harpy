package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) Reconcile(ctx context.Context,
	request adoptionapp.ReconcileRequest,
) (adoptionapp.Run, error) {
	if !storage.usable(ctx) || !storage.owns(request.Authority) ||
		request.FinalLSN == "" || request.AppliedLSN == "" || request.OccurredAt.IsZero() {
		return adoptionapp.Run{}, adoptionapp.ErrInvalid
	}
	result, _, err := storage.scanRun(storage.queryRow(ctx, reconcileSQL,
		request.RunID.String(), int64(request.CutoverGeneration),
		request.ManifestSHA256[:], request.FinalLSN, request.AppliedLSN,
		int64(request.PendingJobCount), request.OccurredAt))
	return result, err
}
