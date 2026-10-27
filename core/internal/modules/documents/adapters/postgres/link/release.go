package link

import (
	"context"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Release(ctx context.Context,
	mutation linkapp.Mutation,
) (domain.Link, bool, error) {
	if !storage.usable(ctx) || mutation.Link.TenantID != storage.tenantID ||
		mutation.ActorID != storage.actorID || !mutation.Link.ID.UUIDv7() ||
		!mutation.Source.ID.Valid() || !mutation.Source.VersionID.Valid() ||
		mutation.Source.Owner == "" || mutation.Source.Kind == "" ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.RequestSHA256 == [32]byte{} || mutation.IdempotencyKey == "" {
		return domain.Link{}, false, linkapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, releaseSQL, mutation.Link.ID.String(),
		storage.actorID.String(), mutation.Source.Owner, mutation.Source.Kind,
		mutation.Source.ID.String(), mutation.Source.VersionID.String(),
		mutation.OccurredAt, mutation.OperationID.String(), mutation.IdempotencyKey,
		mutation.RequestSHA256[:], mutation.AuditID.String()))
}
