//go:build integration

package postgres

import "testing"

func TestAnalyticsDashboardExportRouterPreservesNonMaintenanceOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareNonMaintenanceCutoverGuard(t, fixture)

	t.Run("generic current authority", func(t *testing.T) {
		evidence := createNonMaintenanceExternalActive(t, fixture,
			"workflow.delivery.execute", "workflow-provider", "execute")
		receipt := nonMaintenanceSucceededReceipt(t, evidence)
		if _, err := evidence.repository.CompleteExternalJob(
			fixture.ctx, evidence.permit, receipt,
		); err != nil {
			t.Fatalf("complete generic receipt: %v", err)
		}
		assertNonMaintenanceReceiptSucceeded(t, fixture, evidence, receipt)
	})

	t.Run("cutover bound receipt", func(t *testing.T) {
		evidence := createNonMaintenanceExternalActive(t, fixture,
			"documents.adoption.copy", "documents.adoption", "copy")
		bindNonMaintenanceCutover(t, fixture, evidence.jobID)
		receipt := nonMaintenanceSucceededReceipt(t, evidence)
		if _, err := evidence.repository.CompleteExternalJob(
			fixture.ctx, evidence.permit, receipt,
		); err != nil {
			t.Fatalf("complete cutover receipt: %v", err)
		}
		assertNonMaintenanceReceiptSucceeded(t, fixture, evidence, receipt)
	})

	t.Run("cutover forged operation", func(t *testing.T) {
		evidence := createNonMaintenanceExternalActive(t, fixture,
			"documents.adoption.copy", "documents.adoption", "move")
		bindNonMaintenanceCutover(t, fixture, evidence.jobID)
		receipt := nonMaintenanceSucceededReceipt(t, evidence)
		assertAnalyticsMaintenanceStoreRejected(
			t, fixture, evidence, receipt, "P5603",
		)
	})

	t.Run("generic moved authority", func(t *testing.T) {
		evidence := createNonMaintenanceExternalActive(t, fixture,
			"workflow.delivery.archive", "workflow-provider", "archive")
		receipt := nonMaintenanceSucceededReceipt(t, evidence)
		moveAnalyticsMaintenanceAuthority(t, fixture)
		assertAnalyticsMaintenanceStoreRejected(
			t, fixture, evidence, receipt, "P5601",
		)
	})
}
