package rbacchange

import (
	"context"
	_ "embed"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

//go:embed sql/tenant/write_audit.sql
var writeAuditSQL string

func (store *store) WriteAudit(ctx context.Context, scope tenantScope,
	change rbacdomain.Change) error {
	if change.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != uuid.UUID(change.ActorID) {
		return rbacdomain.ErrSink
	}
	bound, err := store.bind(scope)
	if err != nil {
		return err
	}
	var written bool
	err = bound.database.QueryRow(ctx, writeAuditSQL,
		uuid.UUID(change.OperationID), uuid.UUID(change.ActorID),
		uuid.UUID(change.AggregateID), change.RequestSHA256[:],
		change.CommandKey, change.ResultCode, change.Revision, change.Clock,
	).Scan(&written)
	if err != nil || !written {
		return rbacdomain.ErrSink
	}
	return nil
}
