package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) StartPasswordReset(ctx context.Context,
	scope credentialScope,
	command lifecycle.PasswordResetStartCommand) (lifecycle.MutationResult, error) {
	if !credentialUsable(ctx, scope, credentialPasswordReset) ||
		command.Evidence.Validate() != nil || command.Token.Validate() != nil ||
		command.Delivery.Validate(scope.TenantID(), lifecycle.PurposePasswordReset,
			command.Token) != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, token, envelope := command.Evidence, command.Token, command.Delivery.Envelope
	result, err := scanMutation(bound.database.QueryRow(ctx, passwordResetStartSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], token.ID,
		command.IdentifierKind, command.IdentifierSHA256[:], token.Selector[:],
		token.HashKeyID, token.SecretHMAC[:], token.BindingSHA256[:],
		envelope.ArtifactID, envelope.SubjectUserID, envelope.ResourceID,
		command.Delivery.RecipientIdentifierID, command.Delivery.Reference,
		envelope.KeyID, envelope.Nonce[:], envelope.Ciphertext[:], envelope.Tag[:],
		envelope.AADSHA256[:], envelope.CiphertextSHA256[:]))
	if err == lifecycle.ErrNotFound {
		return lifecycle.MutationResult{}, nil
	}
	return result, err
}
