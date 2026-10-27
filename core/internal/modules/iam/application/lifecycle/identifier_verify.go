package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) VerifyIdentifier(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	command IdentifierVerifyCommand,
) (MutationResult, error) {
	if !service.Valid() || !credentialUsable(ctx, scope, command.Evidence,
		platformtx.CredentialEmailVerification) || command.Proof.Validate() != nil {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.VerifyIdentifier(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishCredential(ctx, scope, result,
		credentialChange(command.Evidence, "iam.identifier.verify", "identifier", result))
}
