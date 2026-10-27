//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
)

type analyticsDashboardExportMaintenanceReceipt struct {
	ID                                 uuid.UUID
	Request                            effectsv1.Request
	Kind, Operation, Code              string
	ProviderReference, ResultReference string
	ResultSHA, RequestSHA              string
	CompletedAt                        time.Time
}

func assertAnalyticsDashboardExportMaintenanceReceiptReplay(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn, exportID, jobID string,
	evidence analyticsDashboardExportMaintenanceReceipt,
) {
	t.Helper()
	storeSQL := loadAnalyticsDashboardExportProductionSource(t,
		"platform/async/adapters/postgres/sql/store_external_effect_receipt.sql")
	before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database, exportID, jobID)
	forgedSHA := mutateAnalyticsDashboardExportDigest(evidence.ResultSHA)
	forged := analyticsExportTransaction(t, ctx, connection)
	var exact bool
	err := forged.QueryRowContext(ctx, storeSQL,
		evidence.arguments(forgedSHA)...).Scan(&exact)
	_ = forged.Rollback()
	requireAnalyticsPGCode(t, err, "P6048", "forged terminal maintenance receipt")
	if before != analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		exportID, jobID) {
		t.Fatal("forged terminal maintenance receipt mutated lifecycle evidence")
	}
	replay := analyticsExportTransaction(t, ctx, connection)
	err = replay.QueryRowContext(ctx, storeSQL,
		evidence.arguments(evidence.ResultSHA)...).Scan(&exact)
	if err != nil || !exact {
		_ = replay.Rollback()
		t.Fatalf("canonical terminal maintenance receipt replay = %v: %v", exact, err)
	}
	if err = replay.Commit(); err != nil {
		t.Fatal(err)
	}
	if before != analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		exportID, jobID) {
		t.Fatal("canonical terminal maintenance receipt replay mutated lifecycle evidence")
	}
}

func (evidence analyticsDashboardExportMaintenanceReceipt) arguments(
	resultSHA string,
) []any {
	return []any{evidence.ID, analyticsTenantA, "tenant", analyticsTenantA,
		"analytics", int64(1), "analytics", int64(1), evidence.RequestSHA,
		evidence.Request.EffectID, evidence.Kind, 1, evidence.Kind + ":" +
			evidence.ProviderReference, "analytics.dashboard-export", evidence.Operation,
		evidence.Request.CorrelationID, evidence.Request.CausationID, "succeeded", nil,
		evidence.Code, evidence.ProviderReference, evidence.ResultReference, resultSHA,
		evidence.CompletedAt}
}

func mutateAnalyticsDashboardExportDigest(value string) string {
	prefix := byte('0')
	if len(value) == 64 && value[0] == prefix {
		prefix = '1'
	}
	if len(value) != 64 {
		return value + "0"
	}
	return string(prefix) + value[1:]
}
