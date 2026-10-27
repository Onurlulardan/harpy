package knowledgefile

import (
	"context"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
)

func (storage *store) PrepareAdopt(ctx context.Context,
	value knowledgeapp.AdoptMutation, now time.Time,
) (knowledgeapp.Lease, bool, error) {
	if !storage.usable(ctx) || !validAdopt(value) || now.IsZero() {
		return knowledgeapp.Lease{}, false, knowledgeapi.ErrInvalid
	}
	lease, err := scanLease(storage.queryRow(ctx, prepareAdoptSQL,
		value.LeaseID.String(), storage.actorID.String(), value.ItemID.String(),
		value.VersionID.String(), value.ReferenceID.String(), value.SourceID.String(),
		now.UTC()), storage.tenantID)
	lease, err = storage.result(lease, err)
	return lease, lease.State == "ADOPTED", err
}

func (storage *store) CompleteAdopt(ctx context.Context,
	value knowledgeapp.AdoptMutation, now time.Time,
) (knowledgeapp.Lease, error) {
	if !storage.usable(ctx) || !validAdopt(value) || now.IsZero() {
		return knowledgeapp.Lease{}, knowledgeapi.ErrInvalid
	}
	lease, err := scanLease(storage.queryRow(ctx, completeAdoptSQL,
		value.LeaseID.String(), storage.actorID.String(), value.ItemID.String(),
		value.VersionID.String(), value.ReferenceID.String(), value.SourceID.String(),
		now.UTC()), storage.tenantID)
	return storage.result(lease, err)
}

func validAdopt(value knowledgeapp.AdoptMutation) bool {
	return value.LeaseID.UUIDv7() && value.ItemID.UUIDv7() && value.VersionID.UUIDv7() &&
		value.ReferenceID.UUIDv7() && value.SourceID.UUIDv7() && value.IdempotencyKey != ""
}
