package folder

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type mutationAction func(context.Context, Mutation) (domain.Folder, bool, error)

func applyBoundMutation(ctx context.Context, action mutationAction, mutation Mutation) (domain.Folder, error) {
	result, _, err := action(ctx, mutation)
	if err != nil {
		return domain.Folder{}, mapStoreError(err)
	}
	if result.ID != mutation.FolderID || result.RowVersion <= mutation.ExpectedVersion {
		return domain.Folder{}, ErrUnavailable
	}
	return result, nil
}

func (application *Application) newMutation(actorID, folderID, parentID domain.ID,
	expected uint64, name, key, action string, now time.Time,
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
	binary.BigEndian.PutUint64(revision[:], expected)
	digest := foundation.Digest([]byte(action), actorID[:], folderID[:], parentID[:],
		revision[:], []byte(name))
	return Mutation{OperationID: operationID, AuditID: auditID, ActorID: actorID,
		FolderID: folderID, ParentID: parentID, ExpectedVersion: expected, Name: name,
		IdempotencyKey: key, Action: action, RequestSHA256: digest, OccurredAt: now.UTC()}, nil
}
