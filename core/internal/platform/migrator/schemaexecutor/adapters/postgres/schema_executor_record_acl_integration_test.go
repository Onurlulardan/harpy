//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func TestSchemaExecutorV2SealsFutureRecordACLsOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
			prepareRecordACLCompletionRoutine(t, ctx, environment)
			environment.createTenant(t, ctx, v2ExactTenant, "tenant_record_acl")
			installSchemaSubmissionV2Authority(
				t, ctx, environment, v2ExactTenant, v2ExactTarget,
			)
			plan := v2ExactPlan(t)
			createV2RelatedForm(t, ctx, environment, plan)
			jobID := schemaExecutorIntegrationUUID(884)
			environment.seedPlanV2(t, ctx, jobID, plan)
			applyV2RecordPlan(t, ctx, environment, jobID, plan)
		})
	}
}

func applyV2RecordPlan(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	jobID string,
	plan schemav2.Plan,
) {
	t.Helper()
	_, lease, validated := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(885), schemaExecutorIntegrationUUID(886))
	installSchemaExecutorCompletionFailure(t, ctx, environment, jobID)
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: validated,
	}); err == nil {
		t.Fatal("completion failure did not roll back V2 table and ACLs")
	}
	assertV2RecordCreateRolledBack(t, ctx, environment, plan)
	removeSchemaExecutorCompletionFailure(t, ctx, environment)
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: validated,
	}); err != nil {
		t.Fatalf("apply V2 form create: %v", err)
	}
	form := plan.Steps[5].IdentifierMapping.Form.PhysicalName
	assertV2RecordACL(t, ctx, environment, "tenant_record_acl", form)

	for ordinal := 1; ordinal < 5; ordinal++ {
		_, lease, validated = environment.startPlan(t, ctx,
			schemaExecutorIntegrationUUID(886+ordinal*2),
			schemaExecutorIntegrationUUID(887+ordinal*2))
		if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
			Lease: lease, Plan: validated,
		}); err != nil {
			t.Fatalf("apply V2 step %d: %v", ordinal, err)
		}
	}
	relation := plan.Steps[5].IdentifierMapping.Relations[0].PhysicalName
	assertV2RecordACL(t, ctx, environment, "tenant_record_acl", relation)

	_, lease, validated = environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(896), schemaExecutorIntegrationUUID(897))
	tamperV2ReaderACL(t, ctx, environment, form, "GRANT INSERT")
	if err := schemaExecutorVerifyStorageV2(ctx, environment.pool,
		"tenant_record_acl", "pg_catalog", validated.CurrentStep()); err == nil {
		t.Fatal("verify_storage accepted expanded reader ACL")
	}
	tamperV2ReaderACL(t, ctx, environment, form, "REVOKE INSERT")
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: validated,
	}); err != nil {
		t.Fatalf("apply V2 verify_storage: %v", err)
	}
}

func createV2RelatedForm(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	plan schemav2.Plan,
) {
	t.Helper()
	target := plan.Steps[5].IdentifierMapping.RelatedForms[0].PhysicalName
	_, err := environment.pool.Exec(ctx, `CREATE TABLE tenant_record_acl.`+
		pgIdentifier(target)+` (id uuid PRIMARY KEY)`)
	if err != nil {
		t.Fatal(err)
	}
}

func assertV2RecordCreateRolledBack(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	plan schemav2.Plan,
) {
	t.Helper()
	form := plan.Steps[0].IdentifierMapping.Form.PhysicalName
	var relation *string
	var schemaUsage bool
	if err := environment.pool.QueryRow(ctx, `SELECT
pg_catalog.to_regclass($1)::text,
pg_catalog.has_schema_privilege('harpy_forms_record_reader', $2, 'USAGE')`,
		"tenant_record_acl."+form, "tenant_record_acl",
	).Scan(&relation, &schemaUsage); err != nil {
		t.Fatal(err)
	}
	if relation != nil || schemaUsage {
		t.Fatalf("failed V2 transaction leaked table=%v schema_usage=%t", relation, schemaUsage)
	}
}

func assertV2RecordACL(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	tableName string,
) {
	t.Helper()
	qualified := schemaName + "." + tableName
	var valid bool
	err := environment.pool.QueryRow(ctx, `SELECT
pg_catalog.has_schema_privilege('harpy_forms_record_reader', $1, 'USAGE')
AND NOT pg_catalog.has_schema_privilege('harpy_forms_record_reader', $1, 'CREATE')
AND pg_catalog.has_schema_privilege('harpy_forms_record_writer', $1, 'USAGE')
AND NOT pg_catalog.has_schema_privilege('harpy_forms_record_writer', $1, 'CREATE')
AND pg_catalog.has_table_privilege('harpy_forms_record_reader', $2, 'SELECT')
AND NOT pg_catalog.has_table_privilege('harpy_forms_record_reader', $2, 'INSERT')
AND NOT pg_catalog.has_table_privilege('harpy_forms_record_reader', $2, 'SELECT WITH GRANT OPTION')
AND pg_catalog.has_table_privilege('harpy_forms_record_writer', $2, 'SELECT')
AND pg_catalog.has_table_privilege('harpy_forms_record_writer', $2, 'INSERT')
AND pg_catalog.has_table_privilege('harpy_forms_record_writer', $2, 'UPDATE')
AND pg_catalog.has_table_privilege('harpy_forms_record_writer', $2, 'DELETE')
AND NOT pg_catalog.has_table_privilege('harpy_forms_record_writer', $2, 'TRUNCATE')
AND NOT pg_catalog.has_table_privilege('harpy_runtime', $2, 'SELECT')
AND NOT pg_catalog.has_table_privilege('harpy_v2_runtime', $2, 'SELECT')`,
		schemaName, qualified,
	).Scan(&valid)
	if err != nil || !valid {
		t.Fatalf("record ACL %s valid=%t err=%v", qualified, valid, err)
	}
}

func tamperV2ReaderACL(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tableName string,
	action string,
) {
	t.Helper()
	keyword := " TO "
	if action == "REVOKE INSERT" {
		keyword = " FROM "
	} else if action != "GRANT INSERT" {
		t.Fatal("invalid reader ACL test action")
	}
	_, err := environment.pool.Exec(ctx, action+` ON TABLE tenant_record_acl.`+
		pgIdentifier(tableName)+keyword+`harpy_forms_record_reader`)
	if err != nil {
		t.Fatal(err)
	}
}
