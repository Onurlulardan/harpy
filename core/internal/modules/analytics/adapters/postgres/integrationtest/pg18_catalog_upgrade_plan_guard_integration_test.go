//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func assertAnalyticsCatalogPlanTypeGuards(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	epoch time.Time,
) {
	t.Helper()
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"candidate missing query", deleteAnalyticsPlanField("query")},
		{"candidate missing parameters", deleteAnalyticsPlanField("parameters")},
		{"candidate missing output", deleteAnalyticsPlanField("output")},
		{"pending missing parameters", func(plan map[string]any) {
			makeAnalyticsPlanPending(plan)
			delete(plan, "parameters")
		}},
		{"pending missing output", func(plan map[string]any) {
			makeAnalyticsPlanPending(plan)
			delete(plan, "output")
		}},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := analyticsCatalogRuntimePlanFixture(
				t, fixture, test.mutate,
			)
			_, err := projectAnalyticsCatalogUpgrade(
				ctx, database, connection, invalid,
				fmt.Sprintf("0.0.%d", index+1),
				analyticsUpgradeUUID(900+index),
				analyticsUpgradeUUID(910+index),
				analyticsUpgradeUUID(920+index),
				epoch.Add(time.Duration(index)*time.Microsecond),
			)
			requireAnalyticsPGCode(
				t, err, "22023", "Analytics Catalog incomplete plan",
			)
		})
	}
}

func deleteAnalyticsPlanField(field string) func(map[string]any) {
	return func(plan map[string]any) {
		delete(plan, field)
	}
}

func makeAnalyticsPlanPending(plan map[string]any) {
	plan["status"] = "pending"
	plan["blocker_code"] = "ANALYTICS_QUERY_CONTRACT_UNSUPPORTED"
	delete(plan, "query")
}
