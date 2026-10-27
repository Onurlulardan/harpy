package upload

import (
	"context"
	"encoding/json"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (storage *store) Create(ctx context.Context,
	input uploadapp.CreateMutation,
) (uploadapp.Snapshot, error) {
	if !validCreate(ctx, storage, input) {
		return uploadapp.Snapshot{}, uploadapp.ErrInvalid
	}
	tags, err := json.Marshal(input.Tags)
	if err != nil {
		return uploadapp.Snapshot{}, uploadapp.ErrInvalid
	}
	result, err := scanSnapshot(storage.queryRow(ctx, createSQL,
		input.ID.String(), input.ActorID.String(), input.DocumentID.String(),
		input.CandidateVersionID.String(), nullableID(input.ExpectedCurrentVersionID),
		input.FolderID.String(), input.RequestSHA256[:], input.ExpectedContentSHA256[:],
		input.ExpectedBytes, input.CiphertextUpperBound, input.OriginalName,
		input.DeclaredMIME, input.Description, tags, input.BillingReservationID.String(),
		input.ExpiresAt,
		input.CreatedAt, input.OperationID.String(), input.AuditID.String(),
		input.IdempotencyKey), storage.tenantID, storage.notFound)
	if err == uploadapp.ErrNotFound {
		return uploadapp.Snapshot{}, uploadapp.ErrConflict
	}
	return result, err
}
