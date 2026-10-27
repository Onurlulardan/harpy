package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) AddIdentifier(ctx context.Context, scope tenantScope,
	command lifecycle.IdentifierAddCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, identifierAddSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.ID, command.UserID,
		command.Kind, command.DisplayValue, command.NormalizedValue,
		command.NormalizedSHA256[:], e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
