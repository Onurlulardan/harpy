package dashboardexportartifact

import (
	"context"
	"errors"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func (runtime *Runtime) Reconcile(
	ctx context.Context, draft exportapp.ArtifactDraft,
) (exportapp.Reconciliation, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || draft.ObjectKey == "" ||
		draft.CiphertextBytes < 1 || draft.HeaderSHA256 == [32]byte{} ||
		draft.CiphertextSHA256 == [32]byte{} {
		return exportapp.Reconciliation{}, exportapp.ErrInvalid
	}
	result, err := runtime.objects.ReconcileExact(ctx, artifactobject.ReconcileExpected{
		Key: draft.ObjectKey, CiphertextBytes: draft.CiphertextBytes,
		CiphertextSHA256: draft.CiphertextSHA256, HeaderSHA256: draft.HeaderSHA256})
	if err != nil && !errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		return exportapp.Reconciliation{}, errors.Join(exportapp.ErrUnavailable, err)
	}
	switch result.State {
	case artifactobject.ReconciliationSealed:
		if result.Identity.Key != draft.ObjectKey || result.Identity.Version == "" ||
			result.ProviderRequestID == "" || result.EvidenceSHA256 == [32]byte{} ||
			result.VerifiedAt.IsZero() {
			return exportapp.Reconciliation{}, exportapp.ErrUnavailable
		}
		return exportapp.Reconciliation{State: exportapp.ReconciliationSealed,
			Receipt: exportapp.UploadReceipt{ObjectVersion: result.Identity.Version,
				ProviderRequestID: result.ProviderRequestID,
				UploadedAt:        result.ObjectModifiedAt.UTC()},
			EvidenceSHA256: result.EvidenceSHA256,
			VerifiedAt:     result.VerifiedAt.UTC()}, nil
	case artifactobject.ReconciliationMissing:
		if result.ProviderRequestID == "" || result.EvidenceSHA256 == [32]byte{} ||
			result.VerifiedAt.IsZero() {
			return exportapp.Reconciliation{}, exportapp.ErrUnavailable
		}
		return exportapp.Reconciliation{State: exportapp.ReconciliationMissing,
			Receipt:        exportapp.UploadReceipt{ProviderRequestID: result.ProviderRequestID},
			EvidenceSHA256: result.EvidenceSHA256, VerifiedAt: result.VerifiedAt.UTC()}, nil
	default:
		return exportapp.Reconciliation{State: exportapp.ReconciliationUnknown},
			exportapp.ErrOutcomeUnknown
	}
}
