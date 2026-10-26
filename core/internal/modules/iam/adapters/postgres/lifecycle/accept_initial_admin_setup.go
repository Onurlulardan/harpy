package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) AcceptInitialAdminSetup(
	ctx context.Context,
	scope credentialScope,
	command lifecycle.InitialAdminSetupAcceptCommand,
) (lifecycle.MutationResult, error) {
	if !credentialUsable(ctx, scope, credentialInitialAdminSetup) ||
		command.Evidence.Validate() != nil || command.Proof.Validate() != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	evidence, proof := command.Evidence, command.Proof
	return scanMutation(bound.database.QueryRow(
		ctx,
		initialAdminSetupAcceptSQL,
		scope.TenantID(), evidence.OperationID, evidence.RequestSHA256[:],
		proof.Selector[:], proof.HashKeyID, proof.SecretHMAC[:],
		proof.BindingSHA256[:], command.Salt[:], command.Tag[:],
	))
}
