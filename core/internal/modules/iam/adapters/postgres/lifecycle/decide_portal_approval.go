package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) DecidePortalApproval(ctx context.Context,
	scope tenantScope,
	command lifecycle.PortalDecisionCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	if !command.Decision.Valid() {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, portalApprovalDecideSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.BindingID,
		command.ExpectedRevision, string(command.Decision), e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
