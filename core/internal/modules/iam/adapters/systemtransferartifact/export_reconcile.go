package systemtransferartifact

import (
	"context"
	"errors"
	"fmt"
	"time"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (runtime *Runtime) ReconcileExportArtifact(
	ctx context.Context,
	effect systemtransfer.ExportArtifactEffectStart,
) (systemtransfer.ExportArtifactReconciliation, error) {
	if runtime == nil || !runtimeInputContext(ctx) || effect.Validate() != nil {
		return systemtransfer.ExportArtifactReconciliation{}, errInvalid
	}
	expectedKey := fmt.Sprintf("%s/v1/%s/%d.fcart", exportPrefix,
		effect.Attempt.ExportID, effect.Attempt.AttemptGeneration)
	expectedRequestSHA := exportUploadRequestSHA(expectedKey,
		effect.Draft.CiphertextSHA256, effect.Draft.CiphertextBytes)
	if effect.ObjectKey != expectedKey || effect.RequestSHA256 != expectedRequestSHA {
		return systemtransfer.ExportArtifactReconciliation{}, errInvalid
	}
	reconciled, err := runtime.exports.ReconcileExact(ctx,
		artifactobject.ReconcileExpected{Key: expectedKey,
			CiphertextBytes:  effect.Draft.CiphertextBytes,
			CiphertextSHA256: effect.Draft.CiphertextSHA256,
			HeaderSHA256:     effect.Draft.HeaderSHA256})
	checkedAt := reconciled.VerifiedAt.UTC().Truncate(time.Microsecond)
	if err != nil || reconciled.State == artifactobject.ReconciliationUnknown {
		return systemtransfer.ExportArtifactReconciliation{
			State:             systemtransfer.ExportArtifactUnknown,
			ProviderRequestID: reconciled.ProviderRequestID,
			EvidenceSHA256:    reconciled.EvidenceSHA256, CheckedAt: checkedAt,
		}, errors.Join(errOutcomeUnknown, err)
	}
	if reconciled.State == artifactobject.ReconciliationMissing {
		result := systemtransfer.ExportArtifactReconciliation{
			State:             systemtransfer.ExportArtifactMissing,
			ProviderRequestID: reconciled.ProviderRequestID,
			EvidenceSHA256:    reconciled.EvidenceSHA256, CheckedAt: checkedAt,
		}
		if result.Validate(effect) != nil {
			return systemtransfer.ExportArtifactReconciliation{}, errOutcomeUnknown
		}
		return result, nil
	}
	return runtime.mapSealedExportReconciliation(effect, reconciled, checkedAt)
}

func (runtime *Runtime) mapSealedExportReconciliation(
	effect systemtransfer.ExportArtifactEffectStart,
	reconciled artifactobject.ReconcileResult,
	checkedAt time.Time,
) (systemtransfer.ExportArtifactReconciliation, error) {
	if reconciled.State != artifactobject.ReconciliationSealed ||
		reconciled.Identity.Key != effect.ObjectKey ||
		reconciled.Identity.Version == "" || reconciled.ProviderRequestID == "" ||
		!validEvidenceTime(reconciled.ObjectModifiedAt) ||
		reconciled.ObjectModifiedAt.Before(effect.StartedAt.Add(-time.Second)) ||
		checkedAt.Before(reconciled.ObjectModifiedAt) {
		return systemtransfer.ExportArtifactReconciliation{}, errOutcomeUnknown
	}
	sealedAt := reconciled.ObjectModifiedAt.UTC().Truncate(time.Microsecond)
	artifact := systemtransfer.ExportArtifactMetadata{
		ObjectKey: effect.ObjectKey, ObjectVersion: reconciled.Identity.Version,
		KeyID: effect.Draft.KeyID, ProviderRequestID: reconciled.ProviderRequestID,
		EffectID: effect.EffectID, EffectReceiptID: effect.EffectReceiptID,
		HeaderSHA256:     effect.Draft.HeaderSHA256,
		CiphertextSHA256: effect.Draft.CiphertextSHA256,
		RequestSHA256:    effect.RequestSHA256,
		ResultSHA256: exportUploadResultSHA(effect.ObjectKey,
			reconciled.Identity.Version, reconciled.ProviderRequestID,
			effect.Draft.CiphertextBytes),
		CiphertextBytes:   effect.Draft.CiphertextBytes,
		WrappedDEK:        append([]byte(nil), effect.Draft.WrappedDEK...),
		WrapNonce:         effect.Draft.WrapNonce,
		AttemptGeneration: effect.Attempt.AttemptGeneration, SealedAt: sealedAt,
	}
	result := systemtransfer.ExportArtifactReconciliation{
		State: systemtransfer.ExportArtifactSealed, Artifact: artifact,
		ProviderRequestID: reconciled.ProviderRequestID,
		EvidenceSHA256:    reconciled.EvidenceSHA256, CheckedAt: checkedAt,
	}
	if result.Validate(effect) != nil {
		return systemtransfer.ExportArtifactReconciliation{}, errOutcomeUnknown
	}
	return result, nil
}

var _ systemtransfer.ExportArtifactOutcomeReconciler = (*Runtime)(nil)
