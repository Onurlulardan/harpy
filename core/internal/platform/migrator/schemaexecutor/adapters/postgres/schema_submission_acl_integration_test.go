//go:build integration

package postgres

import (
	"context"
	"testing"
)

func (suite schemaSubmissionSuite) assertDirectAccessDenied(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	var wrapper, resultWrapper, helper, resultHelper, resolver, tableDML bool
	err := suite.environment.pool.QueryRow(ctx, `SELECT
pg_catalog.has_function_privilege(
 'harpy_v2_runtime',
 'submission_a.forms_submit_schema_plan_v1(uuid,text,uuid,bigint,text,text,integer,bytea,text,jsonb)',
 'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_v2_runtime',
 'submission_a.forms_resolve_schema_result_v1(uuid,text,bigint)', 'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_v2_runtime',
 'platform.platform_migrator_submit_schema_plan_v1(name,uuid,uuid,text,uuid,bigint,text,text,integer,bytea,text,jsonb)',
 'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_v2_runtime',
 'platform.platform_migrator_resolve_schema_result_v1(uuid,uuid,text,bigint)',
 'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_v2_runtime',
 'platform.billing_resolve_current_tenant_authority_v1(uuid)', 'EXECUTE'),
pg_catalog.has_table_privilege(
 'harpy_v2_runtime', 'platform.schema_jobs', 'SELECT,INSERT,UPDATE,DELETE')`).
		Scan(&wrapper, &resultWrapper, &helper, &resultHelper, &resolver, &tableDML)
	if err != nil || !wrapper || !resultWrapper || helper || resultHelper ||
		resolver || tableDML {
		t.Fatalf("submission ACL wrappers/helpers/resolver/table = %t/%t/%t/%t/%t/%t err=%v",
			wrapper, resultWrapper, helper, resultHelper, resolver, tableDML, err)
	}
	steps := schemaSubmissionStepsJSON(t, suite.input)
	_, err = suite.runtime.Exec(ctx, `SELECT * FROM
platform.platform_migrator_submit_schema_plan_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb)`,
		"submission_a", suite.tenantA, suite.input.jobID,
		string(suite.input.plan.Resource.Kind), suite.input.plan.Resource.ID,
		suite.input.plan.Resource.Revision, suite.input.plan.ContractVersion,
		suite.input.plan.Compiler.Version, suite.input.plan.Compiler.PostgreSQLMajor,
		suite.input.prepared.CanonicalJSON(), suite.input.prepared.SHA256(), steps)
	requireSchemaSubmissionState(t, err, "42501")
	_, err = suite.runtime.Exec(ctx, `SELECT * FROM
platform.platform_migrator_resolve_schema_result_v1($1,$2,$3,$4)`,
		suite.tenantA, suite.input.jobID, suite.input.prepared.SHA256(), 1)
	requireSchemaSubmissionState(t, err, "42501")
	_, err = suite.runtime.Exec(ctx,
		`SELECT platform.billing_resolve_current_tenant_authority_v1($1)`, suite.tenantA)
	requireSchemaSubmissionState(t, err, "42501")
	_, err = suite.runtime.Exec(ctx,
		`DELETE FROM platform.schema_jobs WHERE job_id=$1`, suite.input.jobID)
	requireSchemaSubmissionState(t, err, "42501")
}
