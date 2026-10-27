//go:build integration

package postgres

import (
	"strings"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestAnalyticsDashboardExportRunningReceiptPathsOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareAnalyticsMaintenancePG18(t, fixture)
	successes := []struct{ kind, state string }{
		{"analytics.dashboard.export.delete", "deleted"},
		{"analytics.dashboard.export.expire", "expired"},
		{"analytics.dashboard.export.expire", "cancelled"},
		{"analytics.dashboard.export.expire", "failed"},
		{"analytics.dashboard.export.expire", "deleted"},
	}
	for _, test := range successes {
		t.Run(test.kind+"/"+test.state, func(t *testing.T) {
			evidence := createAnalyticsMaintenanceActive(t, fixture, test.kind)
			receipt := analyticsMaintenanceSucceededReceipt(t, evidence, test.state)
			if _, err := evidence.repository.CompleteExternalJob(
				fixture.ctx, evidence.permit, receipt,
			); err != nil {
				t.Fatalf("complete running maintenance receipt: %v", err)
			}
			assertAnalyticsMaintenanceReceiptRows(
				t, fixture, evidence, receipt, "succeeded", "succeeded",
			)
		})
	}
	failures := []struct {
		kind, code, class, state, outcome string
		delay                             time.Duration
	}{
		{"analytics.dashboard.export.delete", "ANALYTICS_EXPORT_REJECTED",
			"permanent", "dead", "dead", 0},
		{"analytics.dashboard.export.expire", "ANALYTICS_EXPORT_CONFLICT",
			"conflict", "queued", "requeued", time.Millisecond},
		{"analytics.dashboard.export.delete", "ANALYTICS_EXPORT_UNAVAILABLE",
			"transient", "queued", "requeued", time.Millisecond},
		{"analytics.dashboard.export.expire", "ANALYTICS_EXPORT_RETRY",
			"transient", "queued", "requeued", time.Millisecond},
		{"analytics.dashboard.export.delete", "ANALYTICS_EXPORT_TERMINAL",
			"permanent", "dead", "dead", 0},
	}
	for _, test := range failures {
		t.Run(test.code, func(t *testing.T) {
			evidence := createAnalyticsMaintenanceActive(t, fixture, test.kind)
			receipt := analyticsMaintenanceFailedReceipt(t, evidence, test.code,
				effectsv1.ErrorClass(test.class))
			failure, err := NewJobFailure(jobsv1.ErrorClass(test.class), test.code,
				strings.Repeat("a", 64), test.delay)
			if err != nil {
				t.Fatal(err)
			}
			result, err := evidence.repository.FailExternalJob(
				fixture.ctx, evidence.permit, receipt, failure,
			)
			if err != nil || string(result.State()) != test.state {
				t.Fatalf("fail running receipt state=%s err=%v", result.State(), err)
			}
			assertAnalyticsMaintenanceReceiptRows(
				t, fixture, evidence, receipt, test.state, test.outcome,
			)
		})
	}
}
