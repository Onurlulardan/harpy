package link

import (
	"context"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Register(ctx context.Context,
	mutation linkapp.Mutation,
) (domain.Link, bool, error) {
	value := mutation.Link
	if !storage.usable(ctx) || value.TenantID != storage.tenantID ||
		mutation.ActorID != storage.actorID || !value.ID.UUIDv7() ||
		!value.ItemID.UUIDv7() || value.State != domain.LinkActive ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.RequestSHA256 == [32]byte{} || mutation.IdempotencyKey == "" {
		return domain.Link{}, false, linkapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, registerSQL, value.ID.String(),
		storage.actorID.String(), value.ItemID.String(), value.SourceOwner,
		value.SourceKind, value.SourceID.String(), value.SourceVersionID.String(),
		value.Usage, mutation.OperationID.String(), int64(value.ReferenceEpoch),
		mutation.OccurredAt, mutation.RequestSHA256[:], mutation.IdempotencyKey,
		mutation.AuditID.String()))
}
