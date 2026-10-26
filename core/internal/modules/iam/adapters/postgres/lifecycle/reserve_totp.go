package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ReserveTOTP(ctx context.Context, scope tenantScope,
	command lifecycle.TOTPStartCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, challenge := command.Evidence, command.Challenge
	return scanMutation(bound.database.QueryRow(ctx, totpStartSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], command.AuthenticatorID,
		challenge.ID, command.UserID, command.SessionID, command.DisplayLabel,
		command.KMSKeyID, command.Nonce, command.Ciphertext, command.AuthenticationTag,
		command.AADSHA256[:], challenge.HMACKeyID, challenge.HMAC[:],
		challenge.ContextSHA256[:], e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
