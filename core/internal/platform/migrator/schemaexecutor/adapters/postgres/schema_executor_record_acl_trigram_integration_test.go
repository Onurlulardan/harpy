//go:build integration

package postgres

import (
	"context"
	"testing"
)

func prepareRecordACLCompletionRoutine(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
CREATE FUNCTION platform.platform_async_enqueue_forms_schema_completion_v1(
 p_schema_job_id uuid,p_correlation_id uuid,p_authority_id uuid,
 p_authority_group_id text,p_authority_group_generation bigint,
 p_billing_generation bigint,p_forms_generation bigint)
RETURNS uuid LANGUAGE sql SET search_path = pg_catalog
AS 'SELECT ''018f47f4-7b1b-7000-8000-000000009999''::uuid'`)
	if err != nil {
		t.Fatalf("prepare schema completion routine: %v", err)
	}
}
