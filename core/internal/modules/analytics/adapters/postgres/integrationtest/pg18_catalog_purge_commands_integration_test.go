//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

func callAnalyticsCatalogCommand(
	ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID, operationID, action string,
	referenceEpoch int64, completedAt time.Time,
) error {
	return withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		var ready bool
		key := operationID + ":" + fixture.componentID + ":" + action
		return tx.QueryRowContext(ctx, `SELECT ready
FROM analytics_catalog_component_apply_v2(
 $1,$2,$3,'analytics',1,'analytics.catalog-definition.v1',$4,$5,$6,$7,$8
)`, receiptID, operationID, fixture.componentID, fixture.contentSHA,
			action, referenceEpoch, key, completedAt,
		).Scan(&ready)
	})
}

func callAnalyticsCatalogBlockers(
	ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID, operationID string, completedAt time.Time,
) (json.RawMessage, error) {
	var blockers json.RawMessage
	err := withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT blockers
FROM analytics_catalog_component_blockers_v1($1,$2,$3,1,1,$4)`,
			receiptID, operationID, fixture.componentID, completedAt,
		).Scan(&blockers)
	})
	return blockers, err
}

func placeAnalyticsLegalHold(
	ctx context.Context, connection *sql.Conn,
	holdID string, occurredAt time.Time,
) error {
	return withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		var state string
		return tx.QueryRowContext(ctx, `SELECT state
FROM analytics_legal_hold_place_v1(
 't_analytics_a',$1,$2,decode(repeat('a5',32),'hex'),$3
)`, holdID, analyticsCatalogActorID, occurredAt).Scan(&state)
	})
}

func releaseAnalyticsLegalHold(
	ctx context.Context, connection *sql.Conn,
	holdID string, occurredAt time.Time,
) error {
	return withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		var state string
		return tx.QueryRowContext(ctx, `SELECT state
FROM analytics_legal_hold_release_v1(
 't_analytics_a',$1,$2,1,$3
)`, holdID, analyticsCatalogActorID, occurredAt).Scan(&state)
	})
}

func applyAnalyticsCatalogPurgePlan(
	ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogPurgeFixture,
	endpointReceipt, dashboardReceipt, operationID string,
	completedAt time.Time,
) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return err
	}
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.endpoint, endpointReceipt,
		operationID, "purge", 1, completedAt,
	); err != nil {
		return err
	}
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.dashboard, dashboardReceipt,
		operationID, "purge", 1, completedAt.Add(time.Microsecond),
	); err != nil {
		return err
	}
	return transaction.Commit()
}

func requireAnalyticsCatalogCommand(
	t fatalHelper, ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID, operationID, action string,
	referenceEpoch int64, completedAt time.Time,
) {
	t.Helper()
	if err := callAnalyticsCatalogCommand(
		ctx, connection, fixture, receiptID, operationID, action,
		referenceEpoch, completedAt,
	); err != nil {
		t.Fatalf("Analytics Catalog %s: %v", action, err)
	}
}

type fatalHelper interface {
	Helper()
	Fatalf(string, ...any)
}

func requireAnalyticsCatalogZeroBlockers(
	t fatalHelper, ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID, operationID string, completedAt time.Time,
) {
	t.Helper()
	blockers, err := callAnalyticsCatalogBlockers(
		ctx, connection, fixture, receiptID, operationID, completedAt,
	)
	if err != nil {
		t.Fatalf("Analytics Catalog blockers: %v", err)
	}
	if string(blockers) != "[]" {
		t.Fatalf("Analytics Catalog blockers = %s, want []", blockers)
	}
}

func analyticsCatalogPurgeKey(
	fixture analyticsCatalogProjectionFixture, operationID string,
) string {
	return fmt.Sprintf("%s:%s:purge", operationID, fixture.componentID)
}
