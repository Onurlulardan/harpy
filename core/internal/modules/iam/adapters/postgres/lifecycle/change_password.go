package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ChangePassword(ctx context.Context, scope tenantScope,
	command lifecycle.PasswordChangeCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, passwordChangeSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.UserID,
		command.ExpectedCredentialRevision, command.Salt[:], command.Tag[:],
		e.ActorID, e.PermissionID, e.ExpectedTokenVersion,
		e.ExpectedAuthorizationVersion))
}
