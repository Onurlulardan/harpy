package recoveryanchorpostgres

import (
	"context"
	_ "embed"
	"encoding/json"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/stage_registration.sql
var stageRegistrationSQL string

func (storage *store) StageRegistration(
	ctx context.Context,
	scope tenantScope,
	command anchordomain.RegisterCommand,
) (anchordomain.StagedMutation, error) {
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != command.ActorID || scope.TenantID() != command.TenantID {
		return anchordomain.StagedMutation{}, anchordomain.ErrInvalid
	}
	origins, err := json.Marshal(command.Origins)
	if err != nil {
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
	row := bound.database.QueryRow(ctx, stageRegistrationSQL,
		command.OperationID, command.EnvironmentID, command.TenantID, command.ActorID,
		command.IdempotencyKey, command.RequestSHA256[:], command.Evidence.CredentialID,
		command.Evidence.PublicKeyCOSE, command.Evidence.AAGUID,
		command.Evidence.AttestationFormat, command.Evidence.AttestationType,
		command.RPID, origins, command.OriginsSHA256[:], command.Evidence.SignCount,
		command.OccurredAt)
	return scanStaged(storage, row, command.OperationID, digest)
}
