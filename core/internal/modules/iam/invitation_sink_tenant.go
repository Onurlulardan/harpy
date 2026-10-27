package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (sink *InvitationChangeSink) WriteTenantChange(
	ctx context.Context,
	scope platformtx.TenantScope,
	change invitation.AuditChange,
) error {
	if sink == nil || sink.audit == nil || sink.now == nil ||
		change.Validate() != nil || scope.ActorID() != change.ActorID {
		return invitation.ErrEffects
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return invitation.ErrEffects
	}
	if err := sink.audit.PersistTenant(ctx, scope, change); err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	job, present, err := sink.deliveryJob(change, scope, authority)
	if err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	if present {
		result, enqueueErr := writer.EnqueueJob(ctx, job)
		if enqueueErr != nil || result.ID != change.OperationID.String() {
			return errors.Join(invitation.ErrEffects, enqueueErr)
		}
	}
	envelope, err := sink.tenantEnvelope(change, scope, authority)
	if err != nil {
		return errors.Join(invitation.ErrEffects, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != change.OperationID.String() {
		return errors.Join(invitation.ErrEffects, err)
	}
	return nil
}
