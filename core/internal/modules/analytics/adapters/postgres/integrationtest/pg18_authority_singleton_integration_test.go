//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsCurrentAuthorityRejectsExtraGroupMemberPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "authority-shape-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("shape fixture reconcile = %d/%d, want 2/0",
					reconciled, remaining)
			}
			applyAnalyticsDocuments(t, ctx, database, "t_analytics_a",
				analyticsDocuments(t, origin, "tenant"))
			if _, err := database.ExecContext(ctx, `
INSERT INTO platform.cutover_group_members(
    scope_kind,scope_id,group_id,domain,ordinal
) VALUES ('tenant',$1,'analytics','billing',2)`,
				analyticsTenantA); err != nil {
				t.Fatal(err)
			}

			connection, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_, _ = connection.ExecContext(context.Background(),
					`RESET ROLE; RESET SESSION AUTHORIZATION`)
				_ = connection.Close()
			}()
			configureAnalyticsRuntimeConnection(t, ctx, connection)

			var dashboardCount int
			err = connection.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM t_analytics_a.analytics_dashboard_list_v1(
    't_analytics_a',1,NULL::uuid
)`).Scan(&dashboardCount)
			if err == nil {
				t.Fatal("Analytics dashboard accepted a non-singleton cutover group")
			}
			requireAnalyticsPGCode(t, err, "P6101",
				"Analytics dashboard extra group member")

			reconciled, remaining, err = callAnalyticsReconcile(
				t, ctx, database, true,
			)
			if err == nil {
				t.Fatal("Analytics reconcile accepted a non-singleton cutover group")
			}
			requireAnalyticsPGCode(t, err, "P6101",
				"Analytics reconcile extra group member")
		})
	}
}
