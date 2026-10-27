//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAnalyticsPG18ProtectedCutoverFreshAndAdoptedEvidence(t *testing.T) {
	t.Run("fresh-denied", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		database := startAnalyticsPG18(t, ctx, "fresh-cutover")
		prepareAnalyticsProtectedPG18(t, ctx, database, "fresh")
		authority := prepareAnalyticsCutoverAuthority(t, ctx, database, "fresh")
		empty := analyticsCutoverEvidence{}
		batch := newAnalyticsCutoverBatch(authority, 1, "complete", nil, nil, &empty)
		if _, _, _, err := executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, batch); err == nil {
			t.Fatal("fresh-origin Analytics V1 adoption was accepted")
		}
		assertAnalyticsCutoverCounts(t, ctx, database, authority, 0, 0, 0)
	})

	t.Run("v1-adopted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		database := startAnalyticsPG18(t, ctx, "v1-adopted-cutover")
		prepareAnalyticsProtectedPG18(t, ctx, database, "v1_adopted")
		authority := prepareAnalyticsCutoverAuthority(t, ctx, database, "v1_adopted")
		initial := loadAnalyticsCutoverValue(t, ctx, database)
		first := newAnalyticsCutoverBatch(authority, 1, "initial_copy", nil, &initial, nil)
		count, keyDigest, stateDigest, err := executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, first)
		if err != nil || count != 1 || len(keyDigest) != 64 || len(stateDigest) != 64 {
			t.Fatalf("Analytics initial batch=%d %q %q %v", count, keyDigest, stateDigest, err)
		}
		replayCount, replayKey, replayState, err := executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, first)
		if err != nil || replayCount != count || replayKey != keyDigest ||
			replayState != stateDigest {
			t.Fatalf("Analytics initial replay=%d %q %q %v",
				replayCount, replayKey, replayState, err)
		}
		assertAnalyticsTamperAndIsolation(t, ctx, database, authority, initial)
		updateAnalyticsCutoverSource(t, ctx, database)
		updated := loadAnalyticsCutoverValue(t, ctx, database)
		delta := newAnalyticsCutoverBatch(authority, 2, "delta", &initial, &updated, nil)
		if count, _, _, err = executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, delta); err != nil || count != 1 {
			t.Fatalf("Analytics delta batch=%d %v", count, err)
		}
		quiesceAnalyticsCutover(t, ctx, database, authority)
		execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.domain_cutovers
SET generation=2 WHERE scope_kind='tenant' AND scope_id=$1 AND domain='iam'`,
			authority.tenantID)
		assertAnalyticsFinalDenied(t, ctx, database, authority,
			"dependency authority differs")
		execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.domain_cutovers
SET generation=1 WHERE scope_kind='tenant' AND scope_id=$1 AND domain='iam'`,
			authority.tenantID)
		mutateAnalyticsStage(t, ctx, database, authority,
			[]string{"approval", "evidence_sha256"}, `"`+hexRepeat("ab")+`"`)
		assertAnalyticsFinalDenied(t, ctx, database, authority,
			"release projection differs")
		restoreAnalyticsStage(t, ctx, database, authority, updated)
		mutateAnalyticsStage(t, ctx, database, authority,
			[]string{"legacy_route", "hit_count"}, `1`)
		assertAnalyticsFinalDenied(t, ctx, database, authority,
			"public projection invalid")
		restoreAnalyticsStage(t, ctx, database, authority, updated)
		mutateAnalyticsStage(t, ctx, database, authority,
			[]string{"differential", "evidence_sha256"}, `"`+hexRepeat("ac")+`"`)
		assertAnalyticsFinalDenied(t, ctx, database, authority,
			"differential digest differs")
		restoreAnalyticsStage(t, ctx, database, authority, updated)
		evidence := loadAnalyticsCutoverEvidence(t, ctx, database, authority)
		final := newAnalyticsCutoverBatch(authority, 3, "complete", nil, nil, &evidence)
		if count, _, stateDigest, err = executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, final); err != nil ||
			count != 0 || stateDigest != evidence.TargetRootSHA256 {
			t.Fatalf("Analytics final batch=%d %q %v", count, stateDigest, err)
		}
		if replayCount, replayKey, replayState, err = executeAnalyticsCutoverBatch(
			ctx, database, authority, authority.schema, final); err != nil ||
			replayCount != 0 || len(replayKey) != 64 || replayState != evidence.TargetRootSHA256 {
			t.Fatalf("Analytics final replay=%d %q %q %v",
				replayCount, replayKey, replayState, err)
		}
		assertAnalyticsCutoverCounts(t, ctx, database, authority, 1, 3, 1)
		assertAnalyticsCutoverProjection(t, ctx, database, authority, evidence)
		assertAnalyticsCutoverACL(t, ctx, database, authority)
	})
}

func assertAnalyticsTamperAndIsolation(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority, initial analyticsCutoverValue,
) {
	t.Helper()
	var target map[string]any
	if err := json.Unmarshal([]byte(initial.targetDocument), &target); err != nil {
		t.Fatal(err)
	}
	target["query"].(map[string]any)["row_revision"] = float64(2)
	initial.targetDocument = analyticsCanonicalJSON(t, ctx, database, target)
	digest := sha256.Sum256([]byte(initial.targetDocument))
	initial.targetSHA256 = hex.EncodeToString(digest[:])
	tampered := newAnalyticsCutoverBatch(authority, 1, "initial_copy", nil, &initial, nil)
	if _, _, _, err := executeAnalyticsCutoverBatch(
		ctx, database, authority, authority.schema, tampered); err == nil {
		t.Fatal("Analytics different-hash replay was accepted")
	}
	if _, _, _, err := executeAnalyticsCutoverBatch(
		ctx, database, authority, "t_analytics_b", tampered); err == nil {
		t.Fatal("Analytics cross-tenant batch was accepted")
	}
	var stage int
	if err := database.QueryRowContext(ctx,
		`SELECT count(*) FROM t_analytics_b.analytics_v1_adoption_stage`).Scan(&stage); err != nil || stage != 0 {
		t.Fatalf("Analytics cross-tenant stage=%d err=%v", stage, err)
	}
}

func updateAnalyticsCutoverSource(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE t_analytics_a.openapi_endpoints
SET name='Legacy report refreshed',description='delta',updated_by=$2,
updated_at='2026-07-21T10:06:00.123Z' WHERE id=$1`,
		analyticsCutoverEndpoint, analyticsCutoverRequester)
}

func quiesceAnalyticsCutover(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority,
) {
	t.Helper()
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.cutover_operations
SET state='quiescing',updated_at=clock_timestamp() WHERE operation_id=$1 AND run_id=$2`,
		authority.operationID, authority.runID)
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.cutover_groups
SET state='quiescing',writer='none' WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='analytics'`, authority.tenantID)
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.domain_cutovers
SET state='quiescing',writer='none' WHERE scope_kind='tenant' AND scope_id=$1
AND domain='analytics'`, authority.tenantID)
}

func assertAnalyticsFinalDenied(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority, fragment string,
) {
	t.Helper()
	evidence := loadAnalyticsCutoverEvidence(t, ctx, database, authority)
	batch := newAnalyticsCutoverBatch(authority, 3, "complete", nil, nil, &evidence)
	if _, _, _, err := executeAnalyticsCutoverBatch(
		ctx, database, authority, authority.schema, batch); err == nil ||
		!strings.Contains(err.Error(), fragment) {
		t.Fatalf("Analytics denial %q: %v", fragment, err)
	}
}

func mutateAnalyticsStage(t *testing.T, ctx context.Context, database *sql.DB,
	authority analyticsCutoverAuthority, path []string, value string,
) {
	t.Helper()
	execAnalyticsCutoverSeed(t, ctx, database, `WITH changed AS (
SELECT jsonb_set(target_document,$3::text[],$4::jsonb,false) AS document
FROM t_analytics_a.analytics_v1_adoption_stage
WHERE operation_id=$1 AND run_id=$2 AND legacy_endpoint_id=$5)
UPDATE t_analytics_a.analytics_v1_adoption_stage AS stage
SET target_document=changed.document,
target_sha256=sha256(convert_to(changed.document::text,'UTF8')) FROM changed
WHERE stage.operation_id=$1 AND stage.run_id=$2 AND stage.legacy_endpoint_id=$5`,
		authority.operationID, authority.runID, path, value, analyticsCutoverEndpoint)
}

func restoreAnalyticsStage(t *testing.T, ctx context.Context, database *sql.DB,
	authority analyticsCutoverAuthority, value analyticsCutoverValue,
) {
	t.Helper()
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE t_analytics_a.analytics_v1_adoption_stage
SET target_document=$3::jsonb,target_sha256=decode($4,'hex')
WHERE operation_id=$1 AND run_id=$2`, authority.operationID, authority.runID,
		value.targetDocument, value.targetSHA256)
}
