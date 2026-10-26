package knowledgefile

import (
	"context"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Create(ctx context.Context,
	seed knowledgeapp.StageSeed,
) (knowledgeapp.Lease, error) {
	value := seed.Lease
	if !storage.usable(ctx) || value.State != "PENDING" ||
		value.TenantID != storage.tenantID || value.ActorID != storage.actorID ||
		value.LeaseID != value.UploadID || !value.LeaseID.UUIDv7() ||
		!value.ItemID.UUIDv7() || !value.VersionID.UUIDv7() ||
		value.ExpectedBytes < 1 || value.ExpectedSHA256 == [32]byte{} ||
		value.Filename == "" || value.MIME == "" || value.ExpiresAt.IsZero() ||
		seed.CreatedAt.IsZero() || !value.ExpiresAt.After(seed.CreatedAt) {
		return knowledgeapp.Lease{}, knowledgeapi.ErrInvalid
	}
	result, err := scanLease(storage.queryRow(ctx, createSQL,
		value.LeaseID.String(), storage.actorID.String(), value.ItemID.String(),
		value.VersionID.String(), value.ExpectedBytes, value.ExpectedSHA256[:],
		value.Filename, value.MIME, value.ExpiresAt.UTC(), seed.CreatedAt.UTC()),
		storage.tenantID)
	return storage.result(result, err)
}

func (storage *store) CompleteStage(ctx context.Context, leaseID domain.ID,
	now time.Time,
) (knowledgeapp.Lease, error) {
	if !storage.usable(ctx) || !leaseID.UUIDv7() || now.IsZero() {
		return knowledgeapp.Lease{}, knowledgeapi.ErrInvalid
	}
	result, err := scanLease(storage.queryRow(ctx, completeStageSQL,
		leaseID.String(), storage.actorID.String(), now.UTC()), storage.tenantID)
	return storage.result(result, err)
}

func (storage *store) result(value knowledgeapp.Lease, err error) (knowledgeapp.Lease, error) {
	if err != nil {
		if storage.notFound(err) {
			return knowledgeapp.Lease{}, knowledgeapi.ErrConflict
		}
		return knowledgeapp.Lease{}, knowledgeapi.ErrUnavailable
	}
	return value, nil
}
