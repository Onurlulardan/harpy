package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) ActivateTOTP(
	ctx context.Context, scope platformtx.TenantScope, command TOTPCompleteCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.AuthenticatorID == uuid.Nil || command.Challenge.ID == uuid.Nil ||
		command.Challenge.HMACKeyID == "" || command.Challenge.HMAC == [32]byte{} ||
		command.Challenge.ContextSHA256 == [32]byte{} ||
		command.ExpectedAuthenticatorRevision < 1 || command.ExpectedChallengeRevision < 1 ||
		command.AcceptedTimeStep < 1 {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.ActivateTOTP(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.totp.enrollment.complete", "authenticator", result))
}
