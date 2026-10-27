//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func assertAnalyticsCatalogPurgeBlockers(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture analyticsCatalogPurgeFixture, epoch time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
INSERT INTO t_analytics_a.analytics_catalog_component_blockers(
 component_id,code,digest,active
) VALUES($1,'ANALYTICS_STATIC_POLICY',repeat('c',64),true)`,
		fixture.endpoint.componentID)
	if err != nil {
		t.Fatal(err)
	}
	defer database.ExecContext(context.Background(), `
DELETE FROM t_analytics_a.analytics_catalog_component_blockers
WHERE component_id=$1 AND code='ANALYTICS_STATIC_POLICY'`,
		fixture.endpoint.componentID)
	connection := analyticsRuntimeConnection(t, ctx, database)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	upgrade := analyticsCatalogVersionFixture(
		t, fixture.endpoint, "purge-blocker-candidate", nil,
	)
	candidateOperation := analyticsUpgradeUUID(2101)
	candidateKey := analyticsCatalogProvisionKey(
		candidateOperation, upgrade.componentID,
	)
	candidateRequest := sha256.Sum256([]byte(candidateKey))
	var candidateReady sql.NullTime
	err = transaction.QueryRowContext(ctx, `SELECT ready_at
FROM analytics_catalog_project_v2(
 't_analytics_a',$1,$2,'hr','2.0.0',$3,$4,1,
 'analytics.catalog-definition.v1',$5,'ENDPOINT','hr',$6,$7,1,$8,
 $9,$10,$11,$12,$13,$14,$15
)`, analyticsUpgradeUUID(2102), analyticsCatalogPackageID,
		analyticsUpgradeUUID(2103), upgrade.componentID, upgrade.contentSHA,
		upgrade.slug, upgrade.semanticSHA[:], upgrade.auxiliaryCount,
		upgrade.document, upgrade.runtimePlan, upgrade.planSHA[:],
		analyticsCatalogActorID, candidateKey, candidateRequest[:],
		epoch.Add(10*time.Microsecond),
	).Scan(&candidateReady)
	if err != nil || !candidateReady.Valid {
		t.Fatalf("Analytics purge candidate: ready=%v err=%v", candidateReady, err)
	}
	publicationKey := "analytics-purge-publication"
	publicationRequest := sha256.Sum256([]byte(publicationKey))
	publicationManifest := sha256.Sum256([]byte("analytics-purge-manifest"))
	var publicationState string
	err = transaction.QueryRowContext(ctx, `SELECT lifecycle_state
FROM analytics_publication_create_v1(
 't_analytics_a',$1,$2,$3,$4,$5,'query',$6,'tenant','none',$7,$8
)`, analyticsUpgradeUUID(2110), analyticsCatalogActorID, publicationKey,
		publicationRequest[:], analyticsUpgradeUUID(2111),
		fixture.endpointProjection.publishedID, publicationManifest[:],
		epoch.Add(11*time.Microsecond),
	).Scan(&publicationState)
	if err != nil || publicationState != "draft" {
		t.Fatalf("Analytics purge publication = %s: %v", publicationState, err)
	}
	var referenceState string
	err = transaction.QueryRowContext(ctx, `SELECT state
FROM analytics_reference_register_v1(
 't_analytics_a',$1,$2,'query_version',$3,'workflow','execution',
 $4,$5,'read',$6
)`, analyticsUpgradeUUID(2120), analyticsCatalogActorID,
		fixture.endpointProjection.publishedID, analyticsUpgradeUUID(2121),
		analyticsUpgradeUUID(2122), epoch.Add(12*time.Microsecond),
	).Scan(&referenceState)
	if err != nil || referenceState != "active" {
		t.Fatalf("Analytics purge reference = %s: %v", referenceState, err)
	}
	var holdState string
	err = transaction.QueryRowContext(ctx, `SELECT state
FROM analytics_legal_hold_place_v1(
 't_analytics_a',$1,$2,decode(repeat('b5',32),'hex'),$3
)`, analyticsUpgradeUUID(2130), analyticsCatalogActorID,
		epoch.Add(13*time.Microsecond),
	).Scan(&holdState)
	if err != nil || holdState != "active" {
		t.Fatalf("Analytics purge legal hold = %s: %v", holdState, err)
	}
	retireOperation := analyticsUpgradeUUID(2140)
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.endpoint, analyticsUpgradeUUID(2141),
		retireOperation, "retire", 1, epoch.Add(14*time.Microsecond),
	); err != nil {
		t.Fatalf("Analytics purge blocker retire: %v", err)
	}
	blockerOperation := analyticsUpgradeUUID(2150)
	blockerReceipt := analyticsUpgradeUUID(2151)
	var storedReceipt string
	var blockers json.RawMessage
	err = transaction.QueryRowContext(ctx, `SELECT receipt_id::text,blockers
FROM analytics_catalog_component_blockers_v1($1,$2,$3,1,1,$4)`,
		blockerReceipt, blockerOperation, fixture.endpoint.componentID,
		epoch.Add(15*time.Microsecond),
	).Scan(&storedReceipt, &blockers)
	if err != nil {
		t.Fatalf("Analytics purge blocker snapshot: %v", err)
	}
	requireAnalyticsPurgeBlockerCodes(t, blockers, []string{
		"ANALYTICS_ACTIVE_INBOUND_REFERENCE",
		"ANALYTICS_CATALOG_CANDIDATE",
		"ANALYTICS_DASHBOARD_WIDGET_REFERENCE",
		"ANALYTICS_LEGAL_HOLD",
		"ANALYTICS_PUBLICATION_RELEASE_REFERENCE",
		"ANALYTICS_STATIC_POLICY",
	})
	var replayReceipt string
	err = transaction.QueryRowContext(ctx, `SELECT receipt_id::text,blockers
FROM analytics_catalog_component_blockers_v1($1,$2,$3,1,1,$4)`,
		analyticsUpgradeUUID(2152), blockerOperation,
		fixture.endpoint.componentID, epoch.Add(16*time.Microsecond),
	).Scan(&replayReceipt, &blockers)
	if err != nil || replayReceipt != storedReceipt || storedReceipt != blockerReceipt {
		t.Fatalf("Analytics blocker replay receipt = %s/%s: %v",
			storedReceipt, replayReceipt, err)
	}
}

func queryAnalyticsCatalogCommandTx(
	ctx context.Context, transaction *sql.Tx,
	fixture analyticsCatalogProjectionFixture,
	receiptID, operationID, action string,
	referenceEpoch int64, completedAt time.Time,
) error {
	var ready bool
	return transaction.QueryRowContext(ctx, `SELECT ready
FROM analytics_catalog_component_apply_v2(
 $1,$2,$3,'analytics',1,'analytics.catalog-definition.v1',$4,$5,$6,$7,$8
)`, receiptID, operationID, fixture.componentID, fixture.contentSHA,
		action, referenceEpoch,
		operationID+":"+fixture.componentID+":"+action, completedAt,
	).Scan(&ready)
}

func requireAnalyticsPurgeBlockerCodes(
	t *testing.T, raw json.RawMessage, expected []string,
) {
	t.Helper()
	var blockers []struct{ Code string }
	if err := json.Unmarshal(raw, &blockers); err != nil {
		t.Fatal(err)
	}
	actual := make(map[string]bool, len(blockers))
	for _, blocker := range blockers {
		actual[blocker.Code] = true
	}
	for _, code := range expected {
		if !actual[code] {
			t.Fatalf("Analytics purge blockers = %s, missing %s", raw, code)
		}
	}
}
