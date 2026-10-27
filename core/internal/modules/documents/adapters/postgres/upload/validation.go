package upload

import (
	"context"
	"slices"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func validCreate(ctx context.Context, storage *store, input uploadapp.CreateMutation) bool {
	description, tags, _, metadataOK := foundation.NormalizeItemMetadata(input.Description,
		input.Tags)
	return storage.usable(ctx) && storage.command && input.TenantID == storage.tenantID &&
		input.ActorID == storage.actorID && input.ID.UUIDv7() && input.DocumentID.UUIDv7() &&
		input.CandidateVersionID.UUIDv7() && input.FolderID.UUIDv7() &&
		input.BillingReservationID.UUIDv7() && input.OperationID.UUIDv7() &&
		input.AuditID.UUIDv7() && input.RequestSHA256 != [32]byte{} &&
		input.ExpectedContentSHA256 != [32]byte{} && input.ExpectedBytes > 0 &&
		input.CiphertextUpperBound > input.ExpectedBytes && input.OriginalName != "" &&
		input.DeclaredMIME != "" && input.State == domain.UploadCreated && metadataOK &&
		description == input.Description && slices.Equal(tags, input.Tags) &&
		foundation.ValidIdempotency(input.IdempotencyKey) && validTime(input.CreatedAt) &&
		validTime(input.UpdatedAt) && validTime(input.ExpiresAt)
}

func validTime(value time.Time) bool { return !value.IsZero() && value.Location() != nil }

func nullableID(value domain.ID) any {
	if !value.Valid() {
		return nil
	}
	return value.String()
}
