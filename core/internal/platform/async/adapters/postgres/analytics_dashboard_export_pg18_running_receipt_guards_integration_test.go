//go:build integration

package postgres

import (
	"strings"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
)

func TestAnalyticsDashboardExportRunningReceiptGuardsOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareAnalyticsMaintenancePG18(t, fixture)

	t.Run("forged receipt identity", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "deleted")
		receipt.ReceiptID = uuid.NewSHA1(uuid.NameSpaceOID,
			[]byte("forged-running-receipt")).String()
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6048")
	})

	t.Run("forged result digest", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.expire")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "expired")
		receipt.ResultSHA256 = analyticsMaintenanceStringPointer(strings.Repeat("f", 64))
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6048")
	})

	t.Run("success with failure code", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "deleted")
		receipt.Code = "ANALYTICS_EXPORT_REJECTED"
		receipt.ReceiptID = analyticsMaintenanceCanonicalReceiptID(
			evidence.permit.Request(), receipt.Code)
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6048")
	})

	t.Run("request digest differs", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "deleted")
		tamperAnalyticsMaintenanceAttemptRequest(t, fixture, evidence.attemptID)
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6048")
	})

	t.Run("stale lease", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.expire")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "expired")
		setAnalyticsDashboardExportLease(t, fixture, evidence.jobID,
			time.Now().UTC().Add(-time.Second))
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6047")
	})

	t.Run("terminal without existing receipt", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.expire")
		setAnalyticsMaintenanceTerminal(t, fixture, evidence, "succeeded")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "expired")
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "P6048")
	})

	t.Run("moved authority", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "deleted")
		moveAnalyticsMaintenanceAuthority(t, fixture)
		assertAnalyticsMaintenanceStoreRejected(t, fixture, evidence, receipt, "42501")
	})
}

func assertAnalyticsMaintenanceStoreRejected(t *testing.T,
	fixture analyticsDashboardExportPG18, evidence analyticsMaintenanceActive,
	receipt effectsv1.Receipt, wantState string,
) {
	t.Helper()
	before := analyticsMaintenanceReceiptSnapshot(t, fixture, evidence.jobID)
	err := storeAnalyticsMaintenanceReceipt(t, fixture, evidence, receipt)
	after := analyticsMaintenanceReceiptSnapshot(t, fixture, evidence.jobID)
	if analyticsExportSQLState(err) != wantState || before != after {
		t.Fatalf("receipt guard state=%s want=%s changed=%v err=%v",
			analyticsExportSQLState(err), wantState, before != after, err)
	}
}

func analyticsMaintenanceStringPointer(value string) *string { return &value }

func tamperAnalyticsMaintenanceAttemptRequest(t *testing.T,
	fixture analyticsDashboardExportPG18, attemptID uuid.UUID,
) {
	t.Helper()
	tx, err := fixture.database.BeginTx(fixture.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(fixture.ctx,
		"ALTER TABLE platform.job_attempts DISABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(fixture.ctx, `UPDATE platform.job_attempts
SET effect_request_sha256=repeat('f',64) WHERE attempt_id=$1`, attemptID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(fixture.ctx,
		"ALTER TABLE platform.job_attempts ENABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func moveAnalyticsMaintenanceAuthority(t *testing.T, fixture analyticsDashboardExportPG18) {
	t.Helper()
	for _, statement := range []string{
		`UPDATE platform.cutover_groups SET state='v1',writer='v1'
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'`,
		`UPDATE platform.domain_cutovers SET state='v1',writer='v1'
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='analytics'`,
	} {
		if _, err := fixture.database.ExecContext(fixture.ctx, statement, asyncTenantA); err != nil {
			t.Fatal(err)
		}
	}
}
