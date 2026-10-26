package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) StartPasswordReset(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	command PasswordResetStartCommand,
) (MutationResult, error) {
	if !service.Valid() || !credentialUsable(ctx, scope, command.Evidence,
		platformtx.CredentialPasswordReset) || command.Token.Validate() != nil ||
		(command.IdentifierKind != "email_v1" && command.IdentifierKind != "portal_v1") ||
		command.IdentifierSHA256 == [32]byte{} ||
		command.Delivery.Validate(scope.TenantID(), PurposePasswordReset,
			command.Token) != nil {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.StartPasswordReset(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if result.ID == uuid.Nil {
		return result, nil
	}
	change := credentialChange(command.Evidence, "iam.password_reset.start",
		"one_time_token", result)
	change.DeliveryArtifactRef = command.Delivery.Reference
	return service.finishCredential(ctx, scope, result, change)
}
