package purge

import (
	"context"
	"encoding/json"

	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type deletionWire struct {
	DeletionID string `json:"deletion_id"`
	VersionID  string `json:"version_id"`
}

func (storage *store) Begin(ctx context.Context,
	mutation purgeapp.Mutation,
) (purgeapp.Result, error) {
	if !storage.usable(ctx) || mutation.TenantID != storage.tenantID ||
		!mutation.ItemID.UUIDv7() || !mutation.OwnerID.Valid() ||
		mutation.State != domain.ItemTrashed && mutation.State != domain.ItemPurgeCooling ||
		!mutation.OperationID.UUIDv7() || !mutation.AuditID.UUIDv7() ||
		mutation.OccurredAt.IsZero() || len(mutation.Deletions) == 0 ||
		len(mutation.Deletions) != len(mutation.Versions) {
		return purgeapp.Result{}, purgeapp.ErrInvalid
	}
	tasks := make([]deletionWire, len(mutation.Deletions))
	seen := make(map[domain.ID]struct{}, len(tasks)*2)
	for index, deletion := range mutation.Deletions {
		_, deletionSeen := seen[deletion.DeletionID]
		_, versionSeen := seen[deletion.ID]
		if !deletion.DeletionID.UUIDv7() || !deletion.ID.UUIDv7() ||
			deletionSeen || versionSeen {
			return purgeapp.Result{}, purgeapp.ErrInvalid
		}
		seen[deletion.DeletionID], seen[deletion.ID] = struct{}{}, struct{}{}
		tasks[index] = deletionWire{DeletionID: deletion.DeletionID.String(),
			VersionID: deletion.ID.String()}
	}
	document, err := json.Marshal(tasks)
	if err != nil {
		return purgeapp.Result{}, purgeapp.ErrUnavailable
	}
	var item string
	var pending int
	err = storage.queryRow(ctx, beginSQL, mutation.ItemID.String(),
		mutation.OccurredAt, document, mutation.OperationID.String(),
		mutation.AuditID.String()).Scan(&item, &pending)
	if err != nil {
		if storage.notFound(err) {
			return purgeapp.Result{}, purgeapp.ErrConflict
		}
		return purgeapp.Result{}, purgeapp.ErrUnavailable
	}
	if item != mutation.ItemID.String() || pending != len(tasks) {
		return purgeapp.Result{}, purgeapp.ErrUnavailable
	}
	return purgeapp.Result{ItemID: mutation.ItemID, State: domain.ItemPurging,
		PendingBlobs: pending}, nil
}
