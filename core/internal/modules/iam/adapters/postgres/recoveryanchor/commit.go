package recoveryanchorpostgres

import (
	"context"
	_ "embed"
	"time"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/commit.sql
var commitSQL string

func (storage *store) Commit(
	ctx context.Context,
	scope tenantScope,
	subject anchordomain.Subject,
	operationID anchordomain.ID,
	now time.Time,
) (anchordomain.CommitResult, error) {
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != subject.ActorID || scope.TenantID() != subject.TenantID {
		return anchordomain.CommitResult{}, anchordomain.ErrInvalid
	}
	var result anchordomain.CommitResult
	err = bound.database.QueryRow(ctx, commitSQL,
		operationID, subject.TenantID, subject.ActorID, now).Scan(
		&result.OperationID, &result.AnchorID, &result.Kind,
		&result.AuthorityGeneration, &result.Replayed)
	if err != nil {
		return anchordomain.CommitResult{}, storeError(err, storage.rowNotFound)
	}
	return result, nil
}
