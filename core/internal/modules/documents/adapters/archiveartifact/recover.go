package archiveartifact

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (writer *Writer) Recover(ctx context.Context,
	plan archiveapp.Plan,
) (archiveapp.ArtifactReceipt, error) {
	if writer == nil || !foundation.ExternalEffectsAllowed(ctx) ||
		!validPlan(plan) || plan.ObjectKey == "" || plan.ObjectVersion == "" ||
		plan.ObjectGeneration == 0 || plan.ArtifactPlaintextBytes < 1 ||
		plan.CiphertextBytes <= plan.ArtifactPlaintextBytes ||
		plan.ArtifactSHA256 == [32]byte{} || plan.CiphertextSHA256 == [32]byte{} ||
		plan.KeyID == "" || len(plan.WrappedDEK) < 115 {
		return archiveapp.ArtifactReceipt{}, archiveapp.ErrInvalid
	}
	expected := uploadapp.ObjectExpected{Key: plan.ObjectKey, Version: plan.ObjectVersion,
		Generation: plan.ObjectGeneration, CiphertextBytes: plan.CiphertextBytes,
		CiphertextSHA256: plan.CiphertextSHA256}
	if err := writer.objects.Verify(ctx, expected); err != nil {
		return archiveapp.ArtifactReceipt{}, archiveapp.ErrUnavailable
	}
	return archiveapp.ArtifactReceipt{ObjectKey: plan.ObjectKey,
		ObjectVersion: plan.ObjectVersion, ObjectGeneration: plan.ObjectGeneration,
		PlaintextBytes:  plan.ArtifactPlaintextBytes,
		CiphertextBytes: plan.CiphertextBytes, PlaintextSHA256: plan.ArtifactSHA256,
		CiphertextSHA256: plan.CiphertextSHA256, KeyID: plan.KeyID,
		WrappedDEK: append([]byte(nil), plan.WrappedDEK...),
		WrittenAt:  plan.UpdatedAt}, nil
}
