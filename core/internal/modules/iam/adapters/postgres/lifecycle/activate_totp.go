package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ActivateTOTP(ctx context.Context, scope tenantScope,
	command lifecycle.TOTPCompleteCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, challenge := command.Evidence, command.Challenge
	return scanMutation(bound.database.QueryRow(ctx, totpCompleteSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.AuthenticatorID,
		challenge.ID, challenge.HMACKeyID, challenge.HMAC[:], challenge.ContextSHA256[:],
		command.ExpectedAuthenticatorRevision, command.ExpectedChallengeRevision,
		command.AcceptedTimeStep, e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
