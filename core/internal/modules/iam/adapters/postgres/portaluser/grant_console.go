package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func (storage *store) GrantConsole(
	ctx context.Context, scope tenantScope, command portaluser.RevisionCommand,
) (portaluser.MutationResult, error) {
	if !usable(ctx, scope, commandMode) || !command.Evidence.Valid() ||
		scope.ActorID() != command.Evidence.ActorID {
		return portaluser.MutationResult{}, portaluser.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return portaluser.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, grantConsoleSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.UserID,
		command.ExpectedRevision, command.AccessModeID, command.AssignmentID,
		e.ActorID, e.PermissionID, e.ExpectedTokenVersion,
		e.ExpectedAuthorizationVersion,
	))
}
