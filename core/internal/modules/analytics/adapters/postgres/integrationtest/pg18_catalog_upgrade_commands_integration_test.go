//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

type analyticsCatalogUpgradeProjection struct {
	analyticsCatalogProjectionResult
	projectionID string
	resourceID   string
	publishedID  string
}

func analyticsUpgradeUUID(seed int) string {
	return fmt.Sprintf("019fa000-0000-7000-8000-%012x", seed)
}

func analyticsCatalogProvisionKey(operationID, componentID string) string {
	return operationID + ":" + componentID + ":provision"
}

func projectAnalyticsCatalogUpgrade(
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	packageVersion string,
	releaseID string,
	receiptID string,
	operationID string,
	occurredAt time.Time,
) (analyticsCatalogUpgradeProjection, error) {
	key := analyticsCatalogProvisionKey(operationID, fixture.componentID)
	requestSHA := sha256.Sum256([]byte(key))
	componentKind := "ENDPOINT"
	if fixture.kind == "dashboard" {
		componentKind = "DASHBOARD"
	}
	var result analyticsCatalogUpgradeProjection
	err := withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT receipt_id::text,lifecycle_state,ready_at
FROM analytics_catalog_project_v2(
 't_analytics_a',$1,$2,'hr',$3,$4,$5,1,
 'analytics.catalog-definition.v1',$6,$7,'hr',$8,$9,1,$10,$11,$12,$13,$14,$15,$16,$17
)`, receiptID, analyticsCatalogPackageID, packageVersion, releaseID,
			fixture.componentID, fixture.contentSHA, componentKind, fixture.slug,
			fixture.semanticSHA[:], fixture.auxiliaryCount, fixture.document,
			fixture.runtimePlan, fixture.planSHA[:], analyticsCatalogActorID,
			key, requestSHA[:], occurredAt,
		).Scan(&result.receiptID, &result.lifecycle, &result.readyAt)
	})
	if err != nil {
		return analyticsCatalogUpgradeProjection{}, err
	}
	err = database.QueryRowContext(ctx, `SELECT projection.receipt_id::text,
projection.operational_resource_id::text,projection.published_version_id::text
FROM t_analytics_a.analytics_catalog_projection_attempts AS attempt
JOIN t_analytics_a.analytics_catalog_version_projections AS projection
  ON projection.receipt_id=attempt.projection_receipt_id
WHERE attempt.receipt_id=$1`, result.receiptID).Scan(
		&result.projectionID, &result.resourceID, &result.publishedID,
	)
	return result, err
}

func applyAnalyticsCatalogUpgradeAction(
	ctx context.Context,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	receiptID string,
	operationID string,
	action string,
	completedAt time.Time,
) error {
	return withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		var ready bool
		key := operationID + ":" + fixture.componentID + ":" + action
		err := tx.QueryRowContext(ctx, `SELECT ready
FROM analytics_catalog_component_apply_v2(
 $1,$2,$3,'analytics',1,'analytics.catalog-definition.v1',$4,$5,1,$6,$7
)`, receiptID, operationID, fixture.componentID, fixture.contentSHA,
			action, key, completedAt,
		).Scan(&ready)
		if err != nil {
			return err
		}
		if !ready {
			return fmt.Errorf("Analytics Catalog %s receipt is not ready", action)
		}
		return nil
	})
}

func requireAnalyticsCatalogUpgradeProjection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	packageVersion string,
	releaseID string,
	receiptID string,
	operationID string,
	occurredAt time.Time,
) analyticsCatalogUpgradeProjection {
	t.Helper()
	result, err := projectAnalyticsCatalogUpgrade(
		ctx, database, connection, fixture, packageVersion, releaseID,
		receiptID, operationID, occurredAt,
	)
	if err != nil || result.receiptID != receiptID || !result.readyAt.Valid {
		t.Fatalf("Analytics Catalog projection = %+v: %v", result, err)
	}
	return result
}

func requireAnalyticsCatalogUpgradeAction(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	operationID string,
	action string,
	receiptSeed int,
	completedAt time.Time,
) {
	t.Helper()
	err := applyAnalyticsCatalogUpgradeAction(
		ctx, connection, fixture, analyticsUpgradeUUID(receiptSeed),
		operationID, action, completedAt,
	)
	if err != nil {
		t.Fatalf("Analytics Catalog %s: %v", action, err)
	}
}

func withAnalyticsCatalogRuntimeTx(
	ctx context.Context,
	connection *sql.Conn,
	run func(*sql.Tx) error,
) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return err
	}
	if err = run(transaction); err != nil {
		return err
	}
	return transaction.Commit()
}
