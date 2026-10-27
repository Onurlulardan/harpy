package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) AcceptInitialAdminSetup(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	command InitialAdminSetupAcceptCommand,
) (MutationResult, error) {
	if !service.Valid() || !credentialUsable(ctx, scope, command.Evidence,
		platformtx.CredentialInitialAdminSetup) || command.Proof.Validate() != nil ||
		command.Salt == [16]byte{} || command.Tag == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.AcceptInitialAdminSetup(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishCredential(ctx, scope, result, credentialChange(
		command.Evidence,
		"iam.initial_admin_setup.accept",
		"password_credential",
		result,
	))
}
