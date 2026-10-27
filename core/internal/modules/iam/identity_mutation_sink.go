package iam

import (
	"context"
	"errors"
	"reflect"
	"time"

	identitychange "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identitychange"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type IdentityChangeClock func() time.Time

type IdentityMutationSink interface {
	WriteChange(context.Context, platformtx.TenantScope, changeapp.Change) error
}

type identityMutationSink struct {
	audit    identityports.ChangeAuditStore[platformtx.TenantScope]
	catalogs platformasync.WriteCatalogs
	now      IdentityChangeClock
}

func NewIdentityMutationSink(
	catalogs platformasync.WriteCatalogs,
	now IdentityChangeClock,
) (IdentityMutationSink, error) {
	if now == nil {
		return nil, changeapp.ErrSink
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, changeapp.ErrSink
	}
	audit, err := identitychange.NewStore()
	if err != nil {
		return nil, err
	}
	return &identityMutationSink{audit: audit, catalogs: catalogs, now: now}, nil
}

func (sink *identityMutationSink) WriteChange(
	ctx context.Context,
	scope platformtx.TenantScope,
	change changeapp.Change,
) error {
	if sink == nil || nilIdentityMutationSinkPort(sink.audit) ||
		sink.now == nil || change.Validate() != nil {
		return changeapp.ErrSink
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return changeapp.ErrSink
	}
	record, err := sink.audit.WriteAudit(ctx, scope, change)
	if err != nil {
		return errors.Join(changeapp.ErrSink, err)
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(changeapp.ErrSink, err)
	}
	envelope, err := sink.identityChangeEnvelope(record, scope, authority)
	if err != nil {
		return errors.Join(changeapp.ErrSink, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != uuid.UUID(change.OperationID).String() {
		return errors.Join(changeapp.ErrSink, err)
	}
	return nil
}

func nilIdentityMutationSinkPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

var _ IdentityMutationSink = (*identityMutationSink)(nil)
