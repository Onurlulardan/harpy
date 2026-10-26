package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) PurgeSessions(ctx context.Context,
	scope tenantScope,
	command lifecycle.SessionRevokeAllCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, sessionRevokeAllSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], e.ActorID,
		e.PermissionID, e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
