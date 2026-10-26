package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) ReserveTOTP(
	ctx context.Context, scope platformtx.TenantScope, command TOTPStartCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.AuthenticatorID == uuid.Nil || command.UserID == uuid.Nil ||
		command.SessionID == uuid.Nil || command.DisplayLabel == "" || command.KMSKeyID == "" ||
		len(command.Nonce) != 12 || len(command.Ciphertext) != 32 ||
		len(command.AuthenticationTag) != 16 || command.AADSHA256 == [32]byte{} ||
		command.Challenge.ID == uuid.Nil || command.Challenge.HMACKeyID == "" ||
		command.Challenge.HMAC == [32]byte{} || command.Challenge.ContextSHA256 == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.ReserveTOTP(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.totp.enrollment.start", "authenticator", result))
}
