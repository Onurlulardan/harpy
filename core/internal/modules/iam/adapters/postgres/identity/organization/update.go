package organizationpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/update.sql
var updateOrganizationSQL string

func (storage *store) UpdateOrganization(
	ctx context.Context, scope tenantScope, command identityapp.UpdateOrganizationCommand,
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
		ctx, updateOrganizationSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), command.ExpectedRevision, command.Name, command.Slug,
		nullableID(command.ParentID), command.ReplaceOwners, databaseIDs(command.OwnerIDs),
		databaseIDs(command.OwnerMembershipIDs), databaseID(profile.SchemaID()), profile.Values(),
		digest[:], databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
