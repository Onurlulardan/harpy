package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) RotateRecoveryCodes(ctx context.Context,
	scope tenantScope,
	command lifecycle.RecoveryIssueCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	codes, err := json.Marshal(command.Codes)
	if err != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, recoveryCodesIssueSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.AuthenticatorID,
		command.UserID, command.DisplayLabel, codes, command.HMACKeyID,
		e.ActorID, e.PermissionID, e.ExpectedTokenVersion,
		e.ExpectedAuthorizationVersion))
}
