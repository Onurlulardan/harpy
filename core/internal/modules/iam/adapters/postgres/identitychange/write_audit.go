package identitychange

import (
	"context"
	_ "embed"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/write_audit.sql
var writeAuditSQL string

func (store *store) WriteAudit(
	ctx context.Context,
	scope tenantScope,
	change identitydomain.Change,
) (identitydomain.RecordedChange, error) {
	if change.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != uuid.UUID(change.ActorID) {
		return identitydomain.RecordedChange{}, identitydomain.ErrChangeSink
	}
	bound, err := store.bind(scope)
	if err != nil {
		return identitydomain.RecordedChange{}, err
	}
	record := identitydomain.RecordedChange{Change: change}
	err = bound.database.QueryRow(
		ctx, writeAuditSQL, uuid.UUID(change.OperationID), uuid.UUID(change.ActorID),
		change.ResourceKind, uuid.UUID(change.AggregateID), change.RequestSHA256[:],
		change.CommandKey, change.ResultCode,
	).Scan(&record.Revision, &record.Clock, &record.EventSequence)
	if err != nil || record.Validate() != nil {
		return identitydomain.RecordedChange{}, identitydomain.ErrChangeSink
	}
	return record, nil
}
