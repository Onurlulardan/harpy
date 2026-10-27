package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (sink *InvitationChangeSink) WriteCredentialChange(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	change invitation.AuditChange,
) error {
	if sink == nil || sink.audit == nil || sink.now == nil ||
		change.Validate() != nil || change.CommandKey != "iam.invitation.accept" ||
		platformtx.RequireCredentialPurpose(
			scope, platformtx.CredentialInvitation,
		) != nil {
		return invitation.ErrEffects
	}
	if err := sink.audit.PersistCredential(ctx, scope, change); err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	writer, err := BindCredentialEventSink(
		scope, change.ActorID, sink.catalogs,
	)
	if err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	envelope, err := sink.credentialEnvelope(change, scope)
	if err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	if err := writer.Enqueue(ctx, envelope); err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	return nil
}
