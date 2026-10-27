package knowledgefile

import (
	"context"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
)

func (storage *store) PrepareRelease(ctx context.Context,
	value knowledgeapp.ReleaseMutation,
) (knowledgeapp.Lease, bool, error) {
	if !storage.usable(ctx) || !validRelease(value) {
		return knowledgeapp.Lease{}, false, knowledgeapi.ErrInvalid
	}
	lease, err := scanLease(storage.queryRow(ctx, prepareReleaseSQL,
		value.ItemID.String(), value.ReferenceID.String(), value.SourceID.String()),
		storage.tenantID)
	lease, err = storage.result(lease, err)
	return lease, lease.State == "RELEASED", err
}

func (storage *store) CompleteRelease(ctx context.Context,
	value knowledgeapp.ReleaseMutation, now time.Time,
) (knowledgeapp.Lease, error) {
	if !storage.usable(ctx) || !validRelease(value) || now.IsZero() {
		return knowledgeapp.Lease{}, knowledgeapi.ErrInvalid
	}
	lease, err := scanLease(storage.queryRow(ctx, completeReleaseSQL,
		value.ItemID.String(), value.ReferenceID.String(), value.SourceID.String(),
		now.UTC()), storage.tenantID)
	return storage.result(lease, err)
}

func validRelease(value knowledgeapp.ReleaseMutation) bool {
	return value.ItemID.UUIDv7() && value.ReferenceID.UUIDv7() &&
		value.SourceID.UUIDv7() && value.IdempotencyKey != ""
}
