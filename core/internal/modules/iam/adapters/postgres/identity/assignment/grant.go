package assignmentpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/grant.sql
var grantAssignmentSQL string

func (storage *store) GrantAssignment(
	ctx context.Context, scope tenantScope, command identityapp.GrantAssignmentCommand,
) (identityapp.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return identityapp.MutationResult{}, identityapp.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(
		ctx, grantAssignmentSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), databaseID(command.UserID), databaseID(command.RoleID), string(command.ScopeKind),
		databaseID(command.ScopeID), nullableID(command.OrganizationID),
		nullableID(command.MembershipID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
