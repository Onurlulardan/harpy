//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertAsyncInboxMatrix(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	t.Run("same event across consumer versions", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			eventID := "83000000-0000-4000-8000-000000000001"
			if err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000002",
				eventID, 1, 1, asyncTenantInboxScope(),
			); err != nil {
				t.Fatalf("insert consumer v1 receipt: %v", err)
			}
			if err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000003",
				eventID, 2, 1, asyncTenantInboxScope(),
			); err != nil {
				t.Fatalf("consumer version was not isolated: %v", err)
			}
		})
	})
	t.Run("duplicate event same consumer version", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			eventID := "83000000-0000-4000-8000-000000000004"
			if err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000005",
				eventID, 1, 1, asyncTenantInboxScope(),
			); err != nil {
				t.Fatalf("insert first event receipt: %v", err)
			}
			err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000006",
				eventID, 1, 2, asyncTenantInboxScope(),
			)
			requireAsyncNamedConstraint(t, err, "inbox_event_unique")
		})
	})
	t.Run("duplicate stream sequence", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			if err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000007",
				"83000000-0000-4000-8000-000000000008", 1, 7,
				asyncTenantInboxScope(),
			); err != nil {
				t.Fatalf("insert first sequence receipt: %v", err)
			}
			err := insertAsyncInboxEvent(
				ctx, tx, "83000000-0000-4000-8000-000000000009",
				"83000000-0000-4000-8000-000000000010", 1, 7,
				asyncTenantInboxScope(),
			)
			requireAsyncNamedConstraint(t, err, "inbox_event_sequence_unique")
		})
	})
	assertAsyncInboxCausation(t, ctx, database)
}

func assertAsyncInboxCausation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	t.Run("bootstrap requires causation", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			if err := insertAsyncInboxBootstrap(
				ctx, tx, "83000000-0000-4000-8000-000000000011", asyncCausation,
			); err != nil {
				t.Fatalf("valid bootstrap receipt rejected: %v", err)
			}
		})
	})
	t.Run("bootstrap without causation", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			err := insertAsyncInboxBootstrap(
				ctx, tx, "83000000-0000-4000-8000-000000000012", nil,
			)
			requireAsyncConstraintRejection(t, err)
		})
	})
	t.Run("event receipt rejects causation", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			id := "83000000-0000-4000-8000-000000000013"
			if err := insertAsyncInboxEvent(
				ctx, tx, id, "83000000-0000-4000-8000-000000000014",
				1, 1, asyncTenantInboxScope(),
			); err != nil {
				t.Fatalf("insert event receipt: %v", err)
			}
			_, err := tx.ExecContext(ctx, `
UPDATE platform.inbox SET causation_id = $2 WHERE id = $1`, id, asyncCausation)
			requireAsyncConstraintRejection(t, err)
		})
	})
}
