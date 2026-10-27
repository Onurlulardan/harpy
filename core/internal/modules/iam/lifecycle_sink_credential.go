package iam

import (
	"context"
	"errors"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (sink *LifecycleChangeSink) WriteCredentialChange(ctx context.Context,
	scope platformtx.CredentialTenantScope, change lifecycle.Change) error {
	if sink == nil || sink.sequences == nil || sink.now == nil ||
		change.Validate() != nil {
		return lifecycle.ErrEffects
	}
	sequence, err := sink.sequences.Credential(ctx, scope, change)
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	writer, err := BindCredentialEventSink(scope, change.ActorID, sink.catalogs)
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	job, present, err := sink.deliveryJob(change, scope.TenantID(),
		scope.IAMGeneration(), scope.AuthorityGroupGeneration())
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	if present {
		id, enqueueErr := writer.EnqueueJob(ctx, job)
		if enqueueErr != nil || id != change.OperationID.String() {
			return errors.Join(lifecycle.ErrEffects, enqueueErr)
		}
	}
	envelope, err := sink.envelope(change, scope.TenantID().String(), sequence,
		eventsv1.Authority{ScopeKind: eventsv1.ScopeTenant,
			ScopeID: scope.TenantID().String(), Domain: "iam",
			CutoverGeneration: scope.IAMGeneration(), WriterOrigin: eventsv1.WriterV2,
			GroupID:         "iam_system_resources",
			GroupGeneration: scope.AuthorityGroupGeneration()})
	if err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	if err := writer.Enqueue(ctx, envelope); err != nil {
		return errors.Join(lifecycle.ErrEffects, err)
	}
	return nil
}
