package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) MarkEffectStarted(
	ctx context.Context, write exportapp.EffectStart,
) error {
	if !validArtifactDraft(write.Draft) || !validAttemptIdentity(write.Attempt) ||
		write.EffectID.Version() != 7 || write.EffectRequestSHA256 == [32]byte{} ||
		write.Attempt.Generation != write.Draft.AttemptGeneration || write.StartedAt.IsZero() {
		return exportapp.ErrInvalid
	}
	var persistedAt time.Time
	err := storage.inTenant(ctx, write.Attempt.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		schema, err := workerSchema(boundContext, bound)
		if err != nil {
			return err
		}
		draft := write.Draft
		return bound.queryRow(boundContext, workerEffectSQL, schema,
			write.Attempt.TenantID, write.Attempt.ExportID, write.Attempt.JobID,
			write.Attempt.AttemptID, write.Attempt.Generation, write.Attempt.LeaseOwner,
			write.EffectID, write.EffectRequestSHA256[:], draft.ObjectKey, draft.Envelope.KEKID,
			draft.Envelope.Ciphertext, draft.Envelope.Nonce[:], draft.HeaderSHA256[:],
			draft.CiphertextSHA256[:], draft.PlaintextSHA256[:], draft.CiphertextBytes,
			draft.PlaintextBytes, draft.Rows, draft.ContentType, draft.Filename,
			write.StartedAt).Scan(&persistedAt)
	})
	if err != nil {
		return mapDispatchError(err)
	}
	if !persistedAt.Equal(write.StartedAt) {
		return exportapp.ErrUnavailable
	}
	return nil
}

func validAttemptIdentity(value exportapp.RenderAttemptIdentity) bool {
	return value.TenantID != uuid.Nil && value.ExportID.Version() == 7 &&
		value.JobID.Version() == 7 && value.AttemptID.Version() == 7 &&
		value.LeaseOwner.Version() == 7 && value.Generation >= 1 &&
		value.Generation <= exportapp.MaximumAttempts
}

func workerSchema(ctx context.Context, bound *workerBound) (string, error) {
	if bound == nil || bound.currentSchema == nil {
		return "", exportapp.ErrUnavailable
	}
	schema, err := bound.currentSchema(ctx)
	if err != nil || !workerTenantSchemaPattern.MatchString(schema) {
		return "", exportapp.ErrUnavailable
	}
	return schema, nil
}

func validArtifactDraft(draft exportapp.ArtifactDraft) bool {
	return draft.AttemptGeneration >= 1 && draft.AttemptGeneration <= exportapp.MaximumAttempts &&
		draft.ObjectKey != "" && draft.Envelope.KEKID != "" &&
		len(draft.Envelope.Ciphertext) == 48 && draft.HeaderSHA256 != [32]byte{} &&
		draft.CiphertextSHA256 != [32]byte{} && draft.PlaintextSHA256 != [32]byte{} &&
		draft.CiphertextBytes >= 35 && draft.PlaintextBytes >= 1 && draft.Rows >= 0 &&
		draft.ContentType != "" && draft.Filename != ""
}
