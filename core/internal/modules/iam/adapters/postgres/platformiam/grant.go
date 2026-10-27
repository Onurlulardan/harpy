package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/grant.sql
var grantSQL string

func (storage *store) Grant(
	ctx context.Context, scope tenantScope, command platformdomain.OperatorGrant,
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
	return scanMutation(bound.database.QueryRow(ctx, grantSQL,
		databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(evidence.ActorID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion, evidence.ExpectedPlatformClock,
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.ACR, command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn,
		command.StepUp.AMRTOTP, databaseID(command.TargetUserID),
		command.TargetExpectedAuthorizationVersion,
		databaseID(command.SubjectID), databaseID(command.RoleID),
		databaseID(command.SubjectRoleBindingID), databaseID(command.RolePermissionBindingID),
		command.Capability))
}
