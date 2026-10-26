package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) FinalizeWebAuthn(
	ctx context.Context, scope platformtx.TenantScope, command WebAuthnCompleteCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.AuthenticatorID == uuid.Nil || command.Challenge.ID == uuid.Nil ||
		command.Challenge.HMACKeyID == "" || command.Challenge.HMAC == [32]byte{} ||
		command.Challenge.ContextSHA256 == [32]byte{} || len(command.CredentialID) < 1 ||
		len(command.CredentialID) > 1024 || len(command.PublicKeyCOSE) < 1 ||
		len(command.PublicKeyCOSE) > 4096 || command.RPID == "" ||
		command.RPIDSHA256 == [32]byte{} || command.OriginPolicyVersion < 1 ||
		command.SignCount < 0 ||
		(command.BackupState && !command.BackupEligible) {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.FinalizeWebAuthn(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.webauthn.enrollment.complete", "authenticator", result))
}
