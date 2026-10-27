package recoveryanchorpostgres

import (
	"context"
	_ "embed"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/list.sql
var listSQL string

func (storage *store) List(
	ctx context.Context,
	scope tenantScope,
	subject anchordomain.Subject,
) ([]anchordomain.Anchor, error) {
	bound, err := storage.bind(ctx, scope, snapshotMode)
	if err != nil || scope.ActorID() != subject.ActorID || scope.TenantID() != subject.TenantID {
		return nil, anchordomain.ErrInvalid
	}
	rows, err := bound.database.Query(ctx, listSQL,
		subject.EnvironmentID, subject.TenantID, subject.ActorID)
	if err != nil {
		return nil, storeError(err, storage.rowNotFound)
	}
	defer rows.Close()
	anchors := make([]anchordomain.Anchor, 0)
	for rows.Next() {
		var anchor anchordomain.Anchor
		if err = rows.Scan(&anchor.ID, &anchor.Generation, &anchor.CredentialID,
			&anchor.RPID, &anchor.UserVerification, &anchor.BackupEligible,
			&anchor.BackupState, &anchor.State); err != nil {
			return nil, storeError(err, storage.rowNotFound)
		}
		anchors = append(anchors, anchor)
	}
	if err = rows.Err(); err != nil {
		return nil, storeError(err, storage.rowNotFound)
	}
	return anchors, nil
}
