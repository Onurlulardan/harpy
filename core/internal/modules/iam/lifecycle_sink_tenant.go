package iam

import (
	"context"
	"errors"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (sink *LifecycleChangeSink) WriteTenantChange(ctx context.Context,
	scope platformtx.TenantScope, change lifecycle.Change) error {
	if sink == nil || sink.sequences == nil || sink.now == nil ||
		change.Validate() != nil || scope.ActorID() != change.ActorID {
		return lifecycle.ErrEffects
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return lifecycle.ErrEffects
	}
	sequence, err := sink.sequences.Tenant(ctx, scope, change)
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	job, present, err := sink.deliveryJob(change, scope.TenantID(),
		authority.Generation(), authority.GroupGeneration())
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	if present {
		result, enqueueErr := writer.EnqueueJob(ctx, job)
		if enqueueErr != nil || result.ID != change.OperationID.String() {
			return errors.Join(lifecycle.ErrEffects, enqueueErr)
		}
	}
	envelope, err := sink.envelope(change, scope.TenantID().String(), sequence,
		eventsv1.Authority{ScopeKind: eventsv1.ScopeTenant,
			ScopeID: authority.ScopeID().String(), Domain: authority.Domain(),
			CutoverGeneration: authority.Generation(), WriterOrigin: eventsv1.WriterV2,
			GroupID: authority.GroupID(), GroupGeneration: authority.GroupGeneration()})
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != change.OperationID.String() {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	return nil
}
