package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *deadLetterRepository) List(
	ctx context.Context,
	request deadletter.AuthorizedList,
) (deadletter.Page, error) {
	tenants := request.TenantIDs()
	if ctx == nil || !repository.valid() || !validDeadLetterList(tenants, request) {
		return deadletter.Page{}, ErrDeadLetterInputInvalid
	}
	cursor := request.Cursor()
	var recordedAt any
	var rootID any
	if cursor != nil {
		recordedAt, rootID = cursor.RecordedAt, cursor.RootID
	}
	return withDeadLetterTransaction(
		ctx, repository,
		func(transaction pgx.Tx) (deadletter.Page, error) {
			rows, err := transaction.Query(
				ctx, listDeadLetterRootsSQL, tenants, cursor != nil,
				recordedAt, rootID, request.Limit()+1,
			)
			if err != nil {
				return deadletter.Page{}, fmt.Errorf("list dead-letter roots: %w", err)
			}
			defer rows.Close()
			return scanDeadLetterPage(rows, request.Limit())
		},
	)
}

func validDeadLetterList(
	tenants []uuid.UUID,
	request deadletter.AuthorizedList,
) bool {
	if len(tenants) < 1 || len(tenants) > deadletter.MaximumTenantScope ||
		request.Limit() < 1 || request.Limit() > deadletter.MaximumPageSize {
		return false
	}
	for _, tenantID := range tenants {
		if tenantID == uuid.Nil {
			return false
		}
	}
	cursor := request.Cursor()
	return cursor == nil || (!cursor.RecordedAt.IsZero() && cursor.RootID != uuid.Nil)
}

func scanDeadLetterPage(rows pgx.Rows, limit int) (deadletter.Page, error) {
	entries := make([]deadletter.Entry, 0, limit+1)
	for rows.Next() {
		entry, err := scanDeadLetterEntry(rows)
		if err != nil {
			return deadletter.Page{}, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return deadletter.Page{}, fmt.Errorf("read dead-letter roots: %w", err)
	}
	page := deadletter.Page{Entries: entries}
	if len(entries) > limit {
		last := entries[limit-1]
		page.Entries = entries[:limit]
		page.Next = &deadletter.Cursor{RecordedAt: last.RecordedAt, RootID: last.RootID}
	}
	return page, nil
}

func scanDeadLetterEntry(row pgx.Row) (deadletter.Entry, error) {
	var entry deadletter.Entry
	var consumer *string
	var consumerVersion *int
	var redacted []byte
	err := row.Scan(
		&entry.RootID, &entry.TenantID, &entry.SourceKind, &entry.SourceID,
		&consumer, &consumerVersion, &entry.SourceContract, &entry.SourceVersion,
		&entry.ReasonClass, &entry.ReasonCode, &entry.PayloadAvailable,
		&redacted, &entry.RecordedAt, &entry.PayloadExpiresAt,
		&entry.MetadataExpiresAt, &entry.RetryActionCount,
	)
	if err != nil {
		return deadletter.Entry{}, fmt.Errorf("scan dead-letter root: %w", err)
	}
	if consumer != nil {
		entry.Consumer = *consumer
	}
	if consumerVersion != nil {
		entry.ConsumerVersion = *consumerVersion
	}
	entry.RedactedEnvelope = append(json.RawMessage(nil), redacted...)
	return entry, nil
}
