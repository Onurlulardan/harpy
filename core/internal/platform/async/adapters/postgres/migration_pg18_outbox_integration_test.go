//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func assertAsyncOutboxMatrix(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	cases := []struct {
		name     string
		mutate   func(*asyncOutboxRow)
		accepted bool
	}{
		{name: "canonical object", accepted: true},
		{name: "delivered canonical object", accepted: true, mutate: func(row *asyncOutboxRow) {
			row.delivered = true
		}},
		{name: "delivered redacted", accepted: true, mutate: func(row *asyncOutboxRow) {
			row.delivered, row.redacted = true, true
		}},
		{name: "pending redacted", mutate: func(row *asyncOutboxRow) {
			row.redacted = true
		}},
		{name: "hash mismatch", mutate: func(row *asyncOutboxRow) {
			row.payloadHash = strings.Repeat("0", 64)
		}},
		{name: "non object", mutate: func(row *asyncOutboxRow) {
			row.payload = []byte(`[]`)
			row.payloadHash = asyncSHA256(row.payload)
		}},
		{name: "over payload budget", mutate: func(row *asyncOutboxRow) {
			row.payload = []byte(`{"value":"` + strings.Repeat("a", 65536) + `"}`)
			row.payloadHash = asyncSHA256(row.payload)
		}},
	}
	for index, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
				row := newAsyncOutboxRow(asyncMatrixUUID(82, index+1))
				if test.mutate != nil {
					test.mutate(&row)
				}
				err := insertAsyncOutbox(ctx, tx, row)
				if test.accepted && err != nil {
					t.Fatalf("valid outbox row rejected: %v", err)
				}
				if !test.accepted {
					requireAsyncConstraintRejection(t, err)
				}
			})
		})
	}
	t.Run("contract and business key required", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			row := newAsyncOutboxRow(asyncMatrixUUID(82, 20))
			if err := insertAsyncOutbox(ctx, tx, row); err != nil {
				t.Fatalf("insert outbox row: %v", err)
			}
			_, err := tx.ExecContext(ctx, `
UPDATE platform.outbox SET business_effect_idempotency_key = NULL
WHERE event_id = $1`, row.id)
			requireAsyncConstraintRejection(t, err)
		})
	})
	t.Run("envelope contract is exact", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			row := newAsyncOutboxRow(asyncMatrixUUID(82, 21))
			if err := insertAsyncOutbox(ctx, tx, row); err != nil {
				t.Fatalf("insert outbox row: %v", err)
			}
			_, err := tx.ExecContext(ctx, `
UPDATE platform.outbox SET envelope_contract_version = 'integration-event.v2'
WHERE event_id = $1`, row.id)
			requireAsyncConstraintRejection(t, err)
		})
	})
	assertAsyncOutboxStreamIdentity(t, ctx, database)
}

func assertAsyncOutboxStreamIdentity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	t.Run("stream identity excludes authority", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			first := newAsyncOutboxRow(asyncMatrixUUID(82, 30))
			if err := insertAsyncOutbox(ctx, tx, first); err != nil {
				t.Fatalf("insert first stream event: %v", err)
			}
			second := newAsyncOutboxRow(asyncMatrixUUID(82, 31))
			second.groupID, second.domain = "communications", "notifications"
			requireAsyncNamedConstraint(
				t, insertAsyncOutbox(ctx, tx, second), "outbox_stream_sequence_unique",
			)
		})
	})
}

func asyncMatrixUUID(prefix, suffix int) string {
	return fmt.Sprintf("%02d000000-0000-4000-8000-%012d", prefix, suffix)
}
