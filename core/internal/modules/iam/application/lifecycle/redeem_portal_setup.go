package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) RedeemPortalSetup(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	command PortalSetupCompleteCommand,
) (MutationResult, error) {
	if !service.Valid() || !credentialUsable(ctx, scope, command.Evidence,
		platformtx.CredentialPortalSetup) || command.Proof.Validate() != nil ||
		command.Salt == [16]byte{} || command.Tag == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.RedeemPortalSetup(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishCredential(ctx, scope, result, credentialChange(command.Evidence,
		"iam.portal_setup.complete", "portal_binding", result))
}
