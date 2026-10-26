package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) RevokeSession(ctx context.Context, scope tenantScope,
	command lifecycle.RevisionCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID || command.ExpectedRevision < 1 {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, sessionRevokeOneSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.ID,
		command.ExpectedRevision, e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
