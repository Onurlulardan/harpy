package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/advance_aggregate_sequence.sql
var advanceAggregateSequenceSQL string

//go:embed sql/tenant/insert_audit_entry.sql
var insertAuditEntrySQL string

func (store *Store) AppendAudit(ctx context.Context, auditID, operationID domain.ID,
	action string, actorID, aggregateID domain.ID, version int64) (int64, error) {
	if !auditID.Valid() || !operationID.Valid() || !actorID.Valid() || !aggregateID.Valid() ||
		action == "" || version < 1 {
		return 0, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return 0, err
	}
	var sequence int64
	err := store.database.QueryRow(ctx, advanceAggregateSequenceSQL, action, aggregateID, version).Scan(&sequence)
	if err != nil {
		return 0, err
	}
	if sequence < 1 {
		return 0, application.ErrUnavailable
	}
	if err = executeOne(ctx, store, insertAuditEntrySQL, auditID, operationID, actorID,
		action, aggregateID, version, sequence); err != nil {
		return 0, err
	}
	return sequence, nil
}
