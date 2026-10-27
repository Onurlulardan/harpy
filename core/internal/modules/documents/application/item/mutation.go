package item

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type mutationAction func(context.Context, Mutation) (Snapshot, bool, error)

func applyMutation(ctx context.Context, action mutationAction,
	mutation Mutation,
) (Snapshot, error) {
	result, _, err := action(ctx, mutation)
	if err != nil {
		return Snapshot{}, mapStoreError(err)
	}
	if result.ID != mutation.ItemID || result.RowVersion <= mutation.ExpectedRowVersion {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}

func (application *Application) newMutation(identity foundation.Identity, itemID,
	folderID domain.ID, rowVersion uint64, description string, tagsJSON []byte,
	reason, key, action string, now time.Time,
) (Mutation, error) {
	operationID, err := application.ids.NewUUIDv7()
	if err != nil || !operationID.UUIDv7() {
		return Mutation{}, ErrUnavailable
	}
	auditID, err := application.ids.NewUUIDv7()
	if err != nil || !auditID.UUIDv7() || auditID == operationID {
		return Mutation{}, ErrUnavailable
	}
	var revision [8]byte
	binary.BigEndian.PutUint64(revision[:], rowVersion)
	digest := foundation.Digest([]byte(action), identity.ActorID[:], itemID[:],
		folderID[:], revision[:], []byte(description), tagsJSON, []byte(reason))
	return Mutation{ItemID: itemID, ActorID: identity.ActorID, TenantID: identity.TenantID,
		FolderID: folderID, OperationID: operationID, AuditID: auditID,
		ExpectedRowVersion: rowVersion, Description: description,
		TagsJSON: append([]byte(nil), tagsJSON...), Reason: reason,
		IdempotencyKey: key, RequestSHA256: digest, OccurredAt: now.UTC()}, nil
}
