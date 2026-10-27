package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (storage *WorkerStore) CompleteReady(
	ctx context.Context, write exportapp.ReadyWrite,
) (exportapp.State, error) {
	if !validAttemptIdentity(write.Attempt) || write.EffectID.Version() != 7 ||
		write.EffectReceiptID.Version() != 7 ||
		write.EffectRequestSHA256 == [32]byte{} || write.EffectResultSHA256 == [32]byte{} ||
		write.Receipt.ObjectVersion == "" ||
		write.Receipt.ProviderRequestID == "" || write.Receipt.UploadedAt.IsZero() ||
		write.CompletedAt.IsZero() || write.CompletedAt.Before(write.Receipt.UploadedAt) {
		return "", exportapp.ErrInvalid
	}
	var state string
	err := storage.inTenant(ctx, write.Attempt.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		schema, err := workerSchema(boundContext, bound)
		if err != nil {
			return err
		}
		return bound.queryRow(boundContext, workerCompleteSQL, schema,
			write.Attempt.TenantID, write.Attempt.ExportID, write.Attempt.JobID,
			write.Attempt.AttemptID, write.Attempt.Generation, write.Attempt.LeaseOwner,
			write.EffectID, write.EffectRequestSHA256[:], write.EffectReceiptID,
			write.EffectResultSHA256[:], write.Receipt.ObjectVersion,
			write.Receipt.ProviderRequestID, write.Receipt.UploadedAt,
			write.CompletedAt).Scan(&state)
	})
	if err != nil {
		return "", mapDispatchError(err)
	}
	result := exportapp.State(state)
	if !result.Valid() {
		return "", exportapp.ErrUnavailable
	}
	return result, nil
}
