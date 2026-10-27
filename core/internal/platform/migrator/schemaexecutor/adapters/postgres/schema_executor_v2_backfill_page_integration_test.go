//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	pagev1 "github.com/Onurlulardan/harpy/contracts/formsrecordpage/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func TestSchemaExecutorV2BackfillPageEvidenceResumesOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
			prepareRecordACLCompletionRoutine(t, ctx, environment)
			environment.createTenant(t, ctx, v2BackfillTenant, "tenant_backfill_v2")
			installSchemaSubmissionV2Authority(
				t, ctx, environment, v2BackfillTenant, v2ExactTarget,
			)
			planValue := v2BackfillPlan(t)
			seedV2BackfillRows(t, ctx, environment, planValue)
			jobID := schemaExecutorIntegrationUUID(924)
			environment.seedPlanV2(t, ctx, jobID, planValue)
			_, lease, plan := environment.startPlan(t, ctx,
				schemaExecutorIntegrationUUID(925), schemaExecutorIntegrationUUID(926))

			applyV2BackfillPage(t, ctx, environment, lease, plan, false)
			first := readV2BackfillProgress(t, ctx, environment, jobID)
			if first.rows != 8 || first.bytes > pagev1.MaximumCanonicalBytes ||
				first.bytes < pagev1.MaximumCanonicalBytes-(64*1024) {
				t.Fatalf("first byte page rows=%d bytes=%d", first.rows, first.bytes)
			}
			assertV2BackfillPageDigest(t, ctx, environment, planValue,
				nil, first.checkpoint, first.bytes, first.sha256)

			applyV2BackfillPage(t, ctx, environment, lease, plan, false)
			second := readV2BackfillProgress(t, ctx, environment, jobID)
			if second.rows != 1008 || second.checkpoint == first.checkpoint ||
				second.bytes > pagev1.MaximumCanonicalBytes {
				t.Fatalf("second row page rows=%d bytes=%d", second.rows, second.bytes)
			}
			assertV2BackfillPageDigest(t, ctx, environment, planValue,
				&first.checkpoint, second.checkpoint, second.bytes, second.sha256)

			if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
				Lease: lease, Plan: plan,
			}); err != nil {
				t.Fatalf("resume final V2 backfill page: %v", err)
			}
			assertV2BackfillComplete(t, ctx, environment, planValue, jobID)
		})
	}
}

func applyV2BackfillPage(
	t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	lease schemaexecutor.Lease, plan schemaexecutor.ValidatedPlan, wantComplete bool,
) {
	t.Helper()
	session, err := environment.store.openTenantSession(ctx, lease)
	if err != nil {
		t.Fatal(err)
	}
	complete, err := environment.store.applyBackfillBatch(ctx, session,
		schemaexecutor.ApplyRequest{Lease: lease, Plan: plan})
	session.close(environment.store.cleanupTimeout)
	if err != nil || complete != wantComplete {
		t.Fatalf("V2 backfill page complete=%t err=%v", complete, err)
	}
}

type v2BackfillProgress struct {
	checkpoint string
	rows       int64
	bytes      int64
	sha256     string
}

func readV2BackfillProgress(
	t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, jobID string,
) v2BackfillProgress {
	t.Helper()
	var value v2BackfillProgress
	err := environment.pool.QueryRow(ctx, `SELECT checkpoint_record_id::text,
processed_rows,checkpoint_canonical_bytes,checkpoint_page_sha256
FROM platform.schema_job_steps WHERE job_id=$1 AND ordinal=0`, jobID).Scan(
		&value.checkpoint, &value.rows, &value.bytes, &value.sha256)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
