//go:build integration

package postgres

import (
	"strings"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestAnalyticsDashboardExportRunningReceiptReplayOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareAnalyticsMaintenancePG18(t, fixture)

	t.Run("completion wins forged insert race and exact replay", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		receipt := analyticsMaintenanceSucceededReceipt(t, evidence, "deleted")
		forged := receipt
		forged.ResultSHA256 = analyticsMaintenanceStringPointer(strings.Repeat("e", 64))
		forgedValues, err := validateExternalReceipt(evidence.permit, forged)
		if err != nil {
			t.Fatal(err)
		}
		canonicalConnection := openAsyncRuntimeRole(t, fixture.ctx, fixture.database)
		forgedConnection := openAsyncRuntimeRole(t, fixture.ctx, fixture.database)
		canonicalRepository := &workerRepository{
			database: canonicalConnection, cleanup: 5 * time.Second,
			identity: evidence.repository.identity,
		}
		start := make(chan struct{})
		canonicalDone, forgedDone := make(chan error, 1), make(chan error, 1)
		go func() {
			<-start
			_, completionErr := canonicalRepository.CompleteExternalJob(
				fixture.ctx, evidence.permit, receipt)
			canonicalDone <- completionErr
		}()
		go func() {
			<-start
			forgedDone <- persistAnalyticsMaintenanceReceipt(
				fixture.ctx, forgedConnection, evidence.permit, forged, forgedValues)
		}()
		close(start)
		canonicalErr, forgedErr := <-canonicalDone, <-forgedDone
		if canonicalErr != nil || analyticsExportSQLState(forgedErr) != "P6048" {
			t.Fatalf("completion/forged race canonical=%v forgedState=%s forged=%v",
				canonicalErr, analyticsExportSQLState(forgedErr), forgedErr)
		}
		assertAnalyticsMaintenanceReceiptRows(
			t, fixture, evidence, receipt, "succeeded", "succeeded")
		assertAnalyticsMaintenanceExactReplay(t, fixture, evidence, receipt,
			forgedConnection)
	})

	for _, test := range []struct {
		name, kind, code, class, state, outcome string
		delay                                   time.Duration
	}{
		{"queued conflict", "analytics.dashboard.export.expire",
			"ANALYTICS_EXPORT_CONFLICT", "conflict", "queued", "requeued", time.Millisecond},
		{"dead permanent", "analytics.dashboard.export.delete",
			"ANALYTICS_EXPORT_REJECTED", "permanent", "dead", "dead", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			evidence := createAnalyticsMaintenanceActive(t, fixture, test.kind)
			receipt := analyticsMaintenanceFailedReceipt(t, evidence, test.code,
				effectsv1.ErrorClass(test.class))
			failure, err := NewJobFailure(jobsv1.ErrorClass(test.class), test.code,
				strings.Repeat("b", 64), test.delay)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = evidence.repository.FailExternalJob(
				fixture.ctx, evidence.permit, receipt, failure); err != nil {
				t.Fatal(err)
			}
			assertAnalyticsMaintenanceReceiptRows(
				t, fixture, evidence, receipt, test.state, test.outcome)
			assertAnalyticsMaintenanceExactReplay(t, fixture, evidence, receipt,
				fixture.runtime)
		})
	}
}

func assertAnalyticsMaintenanceExactReplay(t *testing.T,
	fixture analyticsDashboardExportPG18, evidence analyticsMaintenanceActive,
	receipt effectsv1.Receipt, database workerDatabase,
) {
	t.Helper()
	values, err := validateExternalReceipt(evidence.permit, receipt)
	if err != nil {
		t.Fatal(err)
	}
	before := analyticsMaintenanceReceiptSnapshot(t, fixture, evidence.jobID)
	if err = persistAnalyticsMaintenanceReceipt(
		fixture.ctx, database, evidence.permit, receipt, values); err != nil {
		t.Fatalf("exact terminal receipt replay: %v", err)
	}
	after := analyticsMaintenanceReceiptSnapshot(t, fixture, evidence.jobID)
	if before != after {
		t.Fatal("exact terminal receipt replay changed job, attempt, or receipt evidence")
	}
	forged := receipt
	if forged.Status == effectsv1.ReceiptSucceeded {
		forged.ResultSHA256 = analyticsMaintenanceStringPointer(strings.Repeat("d", 64))
	} else {
		class := effectsv1.ErrorPermanent
		if *forged.ErrorClass == effectsv1.ErrorPermanent {
			class = effectsv1.ErrorTransient
		}
		forged.ErrorClass = &class
	}
	forgedValues, err := validateExternalReceipt(evidence.permit, forged)
	if err != nil {
		t.Fatal(err)
	}
	err = persistAnalyticsMaintenanceReceipt(
		fixture.ctx, database, evidence.permit, forged, forgedValues)
	if analyticsExportSQLState(err) != "P6048" ||
		analyticsMaintenanceReceiptSnapshot(t, fixture, evidence.jobID) != before {
		t.Fatalf("forged terminal replay state=%s err=%v",
			analyticsExportSQLState(err), err)
	}
}
