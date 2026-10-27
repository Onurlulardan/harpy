package legalhold

import (
	"context"

	holdapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/legalhold"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Release(ctx context.Context,
	mutation holdapp.Mutation,
) (domain.LegalHold, bool, error) {
	if !storage.usable(ctx) || mutation.Hold.TenantID != storage.tenantID ||
		mutation.ActorID != storage.actorID || !mutation.Hold.ID.UUIDv7() ||
		mutation.Source.Owner == "" || mutation.Source.CaseID == "" ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.RequestSHA256 == [32]byte{} || mutation.IdempotencyKey == "" {
		return domain.LegalHold{}, false, holdapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, releaseSQL, mutation.Hold.ID.String(),
		storage.actorID.String(), mutation.Source.Owner, mutation.Source.CaseID,
		mutation.OccurredAt, mutation.OperationID.String(), mutation.IdempotencyKey,
		mutation.RequestSHA256[:], mutation.AuditID.String()))
}
