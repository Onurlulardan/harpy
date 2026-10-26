package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) EnableWebAuthn(ctx context.Context, scope tenantScope,
	command lifecycle.WebAuthnStartCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, challenge := command.Evidence, command.Challenge
	return scanMutation(bound.database.QueryRow(ctx, webAuthnStartSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.AuthenticatorID,
		challenge.ID, command.UserID, command.SessionID, command.DisplayLabel,
		challenge.HMACKeyID, challenge.HMAC[:], challenge.ContextSHA256[:],
		command.OriginPolicyVersion, command.RPIDSHA256[:], command.OriginSHA256[:],
		command.UserVerificationRequired, e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
