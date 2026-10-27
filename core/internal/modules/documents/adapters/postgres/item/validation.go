package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
)

func validMutation(ctx context.Context, storage *store, input itemapp.Mutation) bool {
	return storage.usable(ctx) && storage.command && input.ActorID == storage.actorID &&
		input.TenantID == storage.tenantID && input.ItemID.UUIDv7() &&
		input.OperationID.UUIDv7() && input.AuditID.UUIDv7() &&
		input.OperationID != input.AuditID && input.ExpectedRowVersion > 0 &&
		foundation.ValidIdempotency(input.IdempotencyKey) &&
		input.RequestSHA256 != [32]byte{} && !input.OccurredAt.IsZero()
}
