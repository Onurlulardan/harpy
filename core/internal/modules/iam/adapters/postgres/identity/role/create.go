package rolepostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/create.sql
var createRoleSQL string

func (storage *store) CreateRole(
	ctx context.Context, scope tenantScope, command identityapp.CreateRoleCommand,
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
	digest := profile.SHA256()
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(
		ctx, createRoleSQL, scope.TenantID(), databaseID(evidence.OperationID),
		evidence.RequestSHA256[:], databaseID(command.ID), string(command.ScopeKind),
		databaseID(command.ScopeID), nullableID(command.OrganizationID), command.Name,
		command.Description, command.Default, databaseID(profile.SchemaID()),
		profile.Values(), digest[:], databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
