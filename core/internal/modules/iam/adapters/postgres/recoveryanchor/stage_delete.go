package recoveryanchorpostgres

import (
	"context"
	_ "embed"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/stage_delete.sql
var stageDeleteSQL string

func (storage *store) StageDelete(
	ctx context.Context,
	scope tenantScope,
	command anchordomain.DeleteCommand,
) (anchordomain.StagedMutation, error) {
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != command.ActorID || scope.TenantID() != command.TenantID {
		return anchordomain.StagedMutation{}, anchordomain.ErrInvalid
	}
	digest, err := storage.protectedAdminDigest(
		ctx,
		bound,
		command.EnvironmentID,
		command.TenantID,
	)
	if err != nil {
		return anchordomain.StagedMutation{}, err
	}
	return scanStaged(storage, bound.database.QueryRow(ctx, stageDeleteSQL,
		command.OperationID, command.EnvironmentID, command.TenantID, command.ActorID,
		command.AnchorID, command.IdempotencyKey, command.RequestSHA256[:],
		command.OccurredAt), command.OperationID, digest)
}
