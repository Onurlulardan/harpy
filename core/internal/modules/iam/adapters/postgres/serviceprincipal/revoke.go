package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func (store *commandStore) Revoke(
	ctx context.Context,
	scope tenantScope,
	command serviceapp.RevokeCommand,
) (serviceapp.MutationResult, error) {
	if command.Validate() != nil || !evidenceScopeUsable(ctx, scope) ||
		scope.ActorID() != command.Evidence.ActorID {
		return serviceapp.MutationResult{}, serviceapp.ErrInvalidGrant
	}
	bound, err := store.bindCommand(scope)
	if err != nil {
		return serviceapp.MutationResult{}, err
	}
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(
		ctx, revokeSQL, evidence.OperationID, evidence.RequestSHA256[:],
		command.GrantID, command.ExpectedRevision, evidence.ActorID,
		evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion, evidence.PurposeSHA256[:],
		evidence.AuthTime, evidence.AMRPassword, evidence.AMRWebAuthn, evidence.AMRTOTP,
	))
}
