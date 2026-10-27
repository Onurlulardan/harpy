//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func TestSchemaExecutorV2BackfillRejectsOversizedRecordOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
			environment.createTenant(t, ctx, v2BackfillTenant, "tenant_backfill_v2")
			installSchemaSubmissionV2Authority(
				t, ctx, environment, v2BackfillTenant, v2ExactTarget,
			)
			planValue := v2BackfillPlan(t)
			seedV2OversizedBackfillRow(t, ctx, environment, planValue)
			jobID := schemaExecutorIntegrationUUID(927)
			environment.seedPlanV2(t, ctx, jobID, planValue)
			_, lease, plan := environment.startPlan(t, ctx,
				schemaExecutorIntegrationUUID(928), schemaExecutorIntegrationUUID(929))

			err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
				Lease: lease, Plan: plan,
			})
			kind, code := schemaexecutor.FaultDetails(err)
			if kind != schemaexecutor.FaultBlocked ||
				code != "schema_backfill_record_too_large" {
				t.Fatalf("oversized fault kind=%s code=%s err=%v", kind, code, err)
			}
			assertV2OversizedProgressUnchanged(t, ctx, environment, planValue, jobID)
		})
	}
}
