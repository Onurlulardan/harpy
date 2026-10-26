package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func (store *commandStore) Grant(
	ctx context.Context,
	scope tenantScope,
	command serviceapp.GrantCommand,
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
		ctx, grantSQL, evidence.OperationID, evidence.RequestSHA256[:],
		command.GrantID, command.PrincipalID, command.ExpectedRevision,
		string(command.Capability), string(command.Scope.Kind), command.Scope.Owner,
		command.Scope.ResourceKind, nullableUUID(command.Scope.ResourceID),
		command.ValidFrom, nullableTime(command.ValidUntil), command.ManifestSHA256[:],
		evidence.ActorID, evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion, evidence.PurposeSHA256[:],
		evidence.AuthTime, evidence.AMRPassword, evidence.AMRWebAuthn, evidence.AMRTOTP,
	))
}
