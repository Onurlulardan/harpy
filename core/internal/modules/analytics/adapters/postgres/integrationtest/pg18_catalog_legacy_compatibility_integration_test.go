//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func assertAnalyticsCatalogLegacyCompatibility(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	epoch, err := time.Parse(time.RFC3339Nano, analyticsCatalogProjectionEpoch)
	if err != nil {
		t.Fatal(err)
	}
	fixture := analyticsEndpointCatalogFixtureFor(
		t, analyticsUpgradeUUID(2000), "hr-employee-directory-compat",
	)
	modernPlan := append([]byte(nil), fixture.runtimePlan...)
	fixture = analyticsCatalogRuntimePlanFixture(
		t, fixture, func(plan map[string]any) {
			delete(plan, "component_id")
			delete(plan, "source_contract")
			delete(plan, "document_sha256")
		},
	)
	receiptID := analyticsUpgradeUUID(2001)
	releaseID := analyticsUpgradeUUID(2002)
	operationID := analyticsUpgradeUUID(2003)
	key := analyticsCatalogProvisionKey(operationID, fixture.componentID)
	legacy, err := projectAnalyticsCatalogFixture(
		ctx, connection, fixture, receiptID, releaseID, key, epoch,
	)
	if err != nil || legacy.receiptID != receiptID || !legacy.readyAt.Valid {
		t.Fatalf("legacy Analytics projection = %+v: %v", legacy, err)
	}
	projection := readAnalyticsLegacyV2Projection(
		t, ctx, database, legacy, fixture, modernPlan, releaseID, operationID,
	)
	assertAnalyticsCatalogStagedInvisible(
		t, ctx, database, connection, fixture, projection,
	)
	requireAnalyticsLegacyCatalogAction(
		t, ctx, connection, fixture, operationID, "stage", 2004,
		epoch.Add(time.Microsecond),
	)
	assertAnalyticsLegacyCandidateStaged(
		t, ctx, database, fixture.componentID, operationID,
	)
	requireAnalyticsLegacyCatalogAction(
		t, ctx, connection, fixture, operationID, "activate", 2005,
		epoch.Add(2*time.Microsecond),
	)
	assertAnalyticsCatalogVisible(t, ctx, connection, fixture, projection)
	head := readAnalyticsCatalogEndpointHead(
		t, ctx, database, fixture.componentID,
	)
	var candidates, activations int
	if err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_candidates
 WHERE component_id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_projection_activations
 WHERE component_id=$1)`, fixture.componentID).Scan(
		&candidates, &activations,
	); err != nil {
		t.Fatal(err)
	}
	if candidates != 0 || activations != 1 ||
		head.projectionID != projection.projectionID ||
		head.packageVersion != "1.0.0" ||
		head.resourceID != projection.resourceID ||
		head.publishedID != projection.publishedID ||
		head.componentState != "ACTIVE" || head.operationalState != "active" {
		t.Fatalf("legacy Analytics v2 authority = %+v candidate/history=%d/%d",
			head, candidates, activations)
	}
}

func readAnalyticsLegacyV2Projection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	legacy analyticsCatalogProjectionResult,
	fixture analyticsCatalogProjectionFixture,
	modernPlan []byte,
	releaseID string,
	operationID string,
) analyticsCatalogUpgradeProjection {
	t.Helper()
	result := analyticsCatalogUpgradeProjection{
		analyticsCatalogProjectionResult: legacy,
	}
	var candidateAttempt, candidateRelease, candidateKey, candidateContent string
	var storedPlan []byte
	var staged bool
	err := database.QueryRowContext(ctx, `SELECT projection.receipt_id::text,
projection.operational_resource_id::text,projection.published_version_id::text,
candidate.attempt_receipt_id::text,candidate.release_id::text,
attempt.idempotency_key,candidate.content_sha256,
candidate.staged_operation_id IS NULL,projection.runtime_plan_bytes
FROM t_analytics_a.analytics_catalog_component_candidates AS candidate
JOIN t_analytics_a.analytics_catalog_version_projections AS projection
 ON projection.receipt_id=candidate.projection_receipt_id
JOIN t_analytics_a.analytics_catalog_projection_attempts AS attempt
 ON attempt.receipt_id=candidate.attempt_receipt_id
WHERE candidate.component_id=$1`, fixture.componentID).Scan(
		&result.projectionID, &result.resourceID, &result.publishedID,
		&candidateAttempt, &candidateRelease, &candidateKey, &candidateContent,
		&staged, &storedPlan,
	)
	if err != nil {
		t.Fatal(err)
	}
	if candidateAttempt != legacy.receiptID || candidateRelease != releaseID ||
		candidateKey != analyticsCatalogProvisionKey(operationID, fixture.componentID) ||
		candidateContent != fixture.contentSHA || !staged {
		t.Fatalf("legacy Analytics v2 candidate = %s/%s/%s/%s/%t",
			candidateAttempt, candidateRelease, candidateKey, candidateContent, staged)
	}
	plan := analyticsCatalogMap(t, json.RawMessage(storedPlan))
	if plan["component_id"] != fixture.componentID ||
		plan["source_contract"] != analyticsCatalogSourceContract ||
		plan["document_sha256"] != fixture.contentSHA {
		t.Fatalf("legacy Analytics enriched plan binding = %#v", plan)
	}
	storedSHA := sha256.Sum256(storedPlan)
	modernSHA := sha256.Sum256(modernPlan)
	if !bytes.Equal(storedPlan, modernPlan) || storedSHA != modernSHA {
		t.Fatal("legacy Analytics enriched plan is not canonical-byte exact")
	}
	return result
}

func assertAnalyticsLegacyCandidateStaged(
	t *testing.T, ctx context.Context, database *sql.DB,
	componentID, operationID string,
) {
	t.Helper()
	var stagedOperation string
	if err := database.QueryRowContext(ctx, `SELECT staged_operation_id::text
FROM t_analytics_a.analytics_catalog_component_candidates
WHERE component_id=$1`, componentID).Scan(&stagedOperation); err != nil {
		t.Fatal(err)
	}
	if stagedOperation != operationID {
		t.Fatalf("legacy Analytics staged operation = %s", stagedOperation)
	}
}

func requireAnalyticsLegacyCatalogAction(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	operationID, action string,
	receiptSeed int,
	completedAt time.Time,
) {
	t.Helper()
	var ready bool
	err := withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		key := operationID + ":" + fixture.componentID + ":" + action
		return tx.QueryRowContext(ctx, `SELECT ready
FROM analytics_catalog_component_apply_v1(
$1,$2,$3,'analytics',1,'analytics.catalog-definition.v1',$4,$5,1,$6,$7
)`, analyticsUpgradeUUID(receiptSeed), operationID, fixture.componentID,
			fixture.contentSHA, action, key, completedAt,
		).Scan(&ready)
	})
	if err != nil || !ready {
		t.Fatalf("legacy Analytics %s ready=%t: %v", action, ready, err)
	}
}
