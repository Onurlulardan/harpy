package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) RotateRecoveryCodes(
	ctx context.Context, scope platformtx.TenantScope, command RecoveryIssueCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.AuthenticatorID == uuid.Nil || command.UserID == uuid.Nil ||
		command.DisplayLabel == "" || command.HMACKeyID == "" ||
		len(command.Codes) < 8 || len(command.Codes) > 16 {
		return MutationResult{}, ErrInvalid
	}
	seen := make(map[uuid.UUID]struct{}, len(command.Codes))
	for _, code := range command.Codes {
		if code.ID == uuid.Nil || code.HMACSHA256 == "" {
			return MutationResult{}, ErrInvalid
		}
		if _, exists := seen[code.ID]; exists {
			return MutationResult{}, ErrInvalid
		}
		seen[code.ID] = struct{}{}
	}
	result, err := service.store.RotateRecoveryCodes(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.recovery_codes.issue", "authenticator", result))
}
