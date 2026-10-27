package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) RedeemPortalSetup(ctx context.Context,
	scope credentialScope,
	command lifecycle.PortalSetupCompleteCommand) (lifecycle.MutationResult, error) {
	if !credentialUsable(ctx, scope, credentialPortalSetup) ||
		command.Evidence.Validate() != nil || command.Proof.Validate() != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, proof := command.Evidence, command.Proof
	return scanMutation(bound.database.QueryRow(ctx, portalSetupCompleteSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], proof.Selector[:],
		proof.HashKeyID, proof.SecretHMAC[:], proof.BindingSHA256[:],
		command.Salt[:], command.Tag[:]))
}
