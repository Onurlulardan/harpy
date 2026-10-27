package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) CompletePasswordReset(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	command PasswordResetCompleteCommand,
) (MutationResult, error) {
	if !service.Valid() || !credentialUsable(ctx, scope, command.Evidence,
		platformtx.CredentialPasswordReset) || command.Proof.Validate() != nil ||
		command.Salt == [16]byte{} || command.Tag == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.CompletePasswordReset(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishCredential(ctx, scope, result, credentialChange(command.Evidence,
		"iam.password_reset.complete", "password_credential", result))
}
