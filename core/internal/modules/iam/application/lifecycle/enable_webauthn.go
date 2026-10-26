package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) EnableWebAuthn(
	ctx context.Context, scope platformtx.TenantScope, command WebAuthnStartCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.AuthenticatorID == uuid.Nil || command.UserID == uuid.Nil ||
		command.SessionID == uuid.Nil || command.DisplayLabel == "" ||
		command.Challenge.ID == uuid.Nil || command.Challenge.HMACKeyID == "" ||
		command.Challenge.HMAC == [32]byte{} || command.Challenge.ContextSHA256 == [32]byte{} ||
		command.OriginPolicyVersion < 1 || command.RPIDSHA256 == [32]byte{} ||
		command.OriginSHA256 == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.EnableWebAuthn(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.webauthn.enrollment.start", "authenticator", result))
}
