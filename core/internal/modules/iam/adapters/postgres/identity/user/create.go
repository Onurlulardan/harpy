package userpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/create.sql
var createUserSQL string

//go:embed sql/platform/directory_upsert.sql
var directoryUpsertSQL string

func (storage *store) CreateUser(
	ctx context.Context, scope tenantScope, command identityapp.CreateUserCommand,
) (identityapp.MutationResult, error) {
	profile := command.Profile()
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) ||
		profile.SchemaID() == (identityapp.IdentityID{}) {
		return identityapp.MutationResult{}, identityapp.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	authority, err := iamAuthority(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	profileSHA, emailSHA := profile.SHA256(), command.Directory.NormalizedSHA256()
	directoryHMAC := command.Directory.HMAC()
	result, err := scanMutation(bound.database.QueryRow(
		ctx, createUserSQL, scope.TenantID(), databaseID(command.Evidence.OperationID),
		command.Evidence.RequestSHA256[:], databaseID(command.ID), databaseID(command.IdentifierID),
		databaseID(command.AccessModeID), databaseID(command.DefaultAssignmentID), databaseID(profile.SchemaID()),
		profile.Values(), profileSHA[:], command.Directory.Display(), command.Directory.Normalized(),
		emailSHA[:], command.Directory.LookupKeyID(), directoryHMAC[:],
		databaseID(command.Evidence.ActorID), databaseID(command.Evidence.PermissionID),
		command.Evidence.ExpectedTokenVersion,
		command.Evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	var directoryRevision int64
	if err := bound.database.QueryRow(
		ctx, directoryUpsertSQL, scope.TenantID(), authority.GroupGeneration(),
		authority.Generation(), databaseID(command.ID), databaseID(command.IdentifierID),
		command.Directory.LookupKeyID(), directoryHMAC[:], result.Revision, emailSHA[:],
	).Scan(&directoryRevision); err != nil || directoryRevision != result.Revision {
		return identityapp.MutationResult{}, identityapp.ErrStorage
	}
	return result, nil
}
