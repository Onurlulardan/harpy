package permissioncachepostgres

import (
	"context"
	_ "embed"

	permissioncacheapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissioncache"
)

//go:embed sql/tenant/flush.sql
var flushSQL string

func (storage *store) Purge(
	ctx context.Context,
	command permissioncacheapp.Command,
) (permissioncacheapp.Result, error) {
	if !usable(ctx, command.Scope) || command.Scope.ActorID() != command.ActorID ||
		command.Scope.TenantID() != command.TenantID {
		return permissioncacheapp.Result{}, permissioncacheapp.ErrInvalid
	}
	bound, err := storage.bind(command.Scope)
	if err != nil {
		return permissioncacheapp.Result{}, err
	}
	var result permissioncacheapp.Result
	err = bound.database.QueryRow(ctx, flushSQL, command.TenantID, command.OperationID,
		command.RequestSHA256[:], command.ActorID, command.TokenVersion,
		command.AuthorizationVersion, command.PlatformClock).Scan(
		&result.AuthorizationClock, &result.Revision, &result.Replayed)
	if err != nil {
		return permissioncacheapp.Result{}, permissioncacheapp.ErrUnavailable
	}
	return result, nil
}
