//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsFormsReferenceUpgradePostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "forms-reference-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("Analytics authority reconcile = %d/%d", reconciled, remaining)
			}
			applyAnalyticsDocuments(t, ctx, database, "t_analytics_a",
				analyticsDocuments(t, origin, "tenant"))
			var table, claim, ack, releaseClaim, releaseAck, catalogSource *string
			err := database.QueryRowContext(ctx, `
SELECT to_regclass('t_analytics_a.analytics_forms_reference_bindings')::text,
       to_regprocedure('t_analytics_a.analytics_forms_reference_claim_v1(name,uuid,uuid,text,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz)')::text,
       to_regprocedure('t_analytics_a.analytics_forms_reference_register_ack_v1(name,uuid,uuid,bigint,bytea,timestamptz)')::text,
       to_regprocedure('t_analytics_a.analytics_forms_reference_release_claim_v1(name,uuid,uuid,timestamptz)')::text,
       to_regprocedure('t_analytics_a.analytics_forms_reference_release_ack_v1(name,uuid,uuid,bigint,bytea,timestamptz)')::text,
       to_regprocedure('t_analytics_a.analytics_catalog_forms_reference_source_v1(name,uuid)')::text
`).Scan(&table, &claim, &ack, &releaseClaim, &releaseAck, &catalogSource)
			if err != nil || table == nil || claim == nil || ack == nil ||
				releaseClaim == nil || releaseAck == nil || catalogSource == nil {
				t.Fatalf("Analytics Forms reference surface is incomplete: %v", err)
			}
		})
	}
}
