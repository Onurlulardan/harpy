package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) FinalizeWebAuthn(ctx context.Context, scope tenantScope,
	command lifecycle.WebAuthnCompleteCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, challenge := command.Evidence, command.Challenge
	return scanMutation(bound.database.QueryRow(ctx, webAuthnCompleteSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.AuthenticatorID,
		challenge.ID, challenge.HMACKeyID, challenge.HMAC[:], challenge.ContextSHA256[:],
		command.CredentialID, command.PublicKeyCOSE, command.RPID, command.RPIDSHA256[:],
		command.OriginPolicyVersion, command.AAGUID, command.SignCount,
		command.BackupEligible, command.BackupState, command.Discoverable,
		command.UserVerificationRequired, e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
