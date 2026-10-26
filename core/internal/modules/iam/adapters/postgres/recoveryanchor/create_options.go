package recoveryanchorpostgres

import (
	"context"
	_ "embed"
	"encoding/json"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/create_options.sql
var createOptionsSQL string

func (storage *store) CreateOptions(
	ctx context.Context,
	scope tenantScope,
	command anchordomain.OptionsCommand,
) (anchordomain.RegistrationOptions, error) {
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != command.ActorID || scope.TenantID() != command.TenantID {
		return anchordomain.RegistrationOptions{}, anchordomain.ErrInvalid
	}
	origins, err := json.Marshal(command.Origins)
	if err != nil {
		return anchordomain.RegistrationOptions{}, anchordomain.ErrInvalid
	}
	var result anchordomain.RegistrationOptions
	var rawOrigins []byte
	err = bound.database.QueryRow(ctx, createOptionsSQL,
		command.OperationID, command.AnchorID, command.EnvironmentID,
		command.TenantID, command.ActorID, command.IdempotencyKey,
		command.RequestSHA256[:], command.Challenge, command.RPID, origins,
		command.OriginsSHA256[:], command.ExpiresAt, command.OccurredAt,
	).Scan(&result.OperationID, &result.AnchorID, &result.Challenge,
		&result.RPID, &rawOrigins, &result.ExpiresAt, &result.Replayed)
	if err != nil {
		return anchordomain.RegistrationOptions{}, storeError(err, storage.rowNotFound)
	}
	result.Origins, err = decodeOrigins(rawOrigins)
	if err != nil {
		return anchordomain.RegistrationOptions{}, err
	}
	return result, nil
}
