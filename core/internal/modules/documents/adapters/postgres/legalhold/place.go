package legalhold

import (
	"context"

	holdapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/legalhold"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Place(ctx context.Context,
	mutation holdapp.Mutation,
) (domain.LegalHold, bool, error) {
	value := mutation.Hold
	if !storage.usable(ctx) || value.TenantID != storage.tenantID ||
		mutation.ActorID != storage.actorID || !value.ID.UUIDv7() ||
		!value.ItemID.UUIDv7() || value.State != domain.LegalHoldActive ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.RequestSHA256 == [32]byte{} || mutation.IdempotencyKey == "" {
		return domain.LegalHold{}, false, holdapp.ErrInvalid
	}
	return storage.scan(storage.queryRow(ctx, placeSQL, value.ID.String(),
		storage.actorID.String(), value.ItemID.String(), value.SourceOwner,
		value.SourceCaseID, value.EvidenceSHA256[:], mutation.OccurredAt,
		mutation.OperationID.String(), mutation.IdempotencyKey,
		mutation.RequestSHA256[:], mutation.AuditID.String()))
}
