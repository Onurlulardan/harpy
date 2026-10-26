package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) CreatePortalSetup(ctx context.Context, scope tenantScope,
	command lifecycle.PortalSetupStartCommand) (lifecycle.MutationResult, error) {
	if !commandUsable(ctx, scope) || command.Evidence.Validate() != nil ||
		command.Token.Validate() != nil || scope.ActorID() != command.Evidence.ActorID ||
		command.Delivery.Validate(scope.TenantID(), lifecycle.PurposePortalSetup,
			command.Token) != nil {
		return lifecycle.MutationResult{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.MutationResult{}, err
	}
	e, token, envelope := command.Evidence, command.Token, command.Delivery.Envelope
	return scanMutation(bound.database.QueryRow(ctx, portalSetupStartSQL,
		scope.TenantID(), e.OperationID, e.RequestSHA256[:], token.ID,
		command.BindingID, token.Selector[:], token.HashKeyID, token.SecretHMAC[:],
		token.BindingSHA256[:], e.ActorID, e.PermissionID,
		e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion,
		envelope.ArtifactID, envelope.SubjectUserID, envelope.ResourceID,
		command.Delivery.RecipientIdentifierID, command.Delivery.Reference,
		envelope.KeyID, envelope.Nonce[:], envelope.Ciphertext[:], envelope.Tag[:],
		envelope.AADSHA256[:], envelope.CiphertextSHA256[:]))
}
