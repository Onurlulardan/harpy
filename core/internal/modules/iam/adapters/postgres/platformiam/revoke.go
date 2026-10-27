package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/revoke.sql
var revokeSQL string

func (storage *store) Revoke(
	ctx context.Context, scope tenantScope, command platformdomain.OperatorRevoke,
) (platformdomain.MutationResult, error) {
	if !commandScopeUsable(ctx, scope) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return platformdomain.MutationResult{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.MutationResult{}, err
	}
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, revokeSQL,
		databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(evidence.ActorID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion, evidence.ExpectedPlatformClock,
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.ACR, command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn,
		command.StepUp.AMRTOTP, databaseID(command.TargetUserID),
		command.TargetExpectedAuthorizationVersion,
		command.Capability))
}
