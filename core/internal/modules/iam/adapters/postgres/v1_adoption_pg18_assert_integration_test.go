//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMV1AdoptionResult(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, operationID, runID, organizationID string,
) {
	t.Helper()
	var state, groupState, writer string
	var organizationCount int
	query := `SELECT
 (SELECT state FROM ` + tenant.schema + `.iam_v1_adoption_runs
   WHERE operation_id=$1 AND run_id=$2),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_organizations WHERE id=$3),
 (SELECT state FROM platform.cutover_groups
   WHERE scope_kind='tenant' AND scope_id=$4 AND group_id='iam_system_resources'),
 (SELECT writer FROM platform.cutover_groups
   WHERE scope_kind='tenant' AND scope_id=$4 AND group_id='iam_system_resources')`
	if err := database.QueryRowContext(
		ctx, query, operationID, runID, organizationID, tenant.id,
	).Scan(&state, &organizationCount, &groupState, &writer); err != nil {
		t.Fatalf("read IAM V1 adoption result in %s: %v", tenant.schema, err)
	}
	if state != "verified" || organizationCount != 1 ||
		groupState != "shadow" || writer != "v1" {
		t.Fatalf("IAM V1 adoption result = %s/%d/%s/%s",
			state, organizationCount, groupState, writer)
	}
	assertIAMV1AdoptedMatrix(t, ctx, database, tenant)
}

func assertIAMV1AdoptionIsolation(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		other := iamFoundationTenants[1-index]
		var leaked int
		query := `SELECT
 (SELECT count(*) FROM ` + tenant.schema + `.iam_organizations WHERE id=$1) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_users WHERE id=$2) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_permissions WHERE id=$3) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_invitations WHERE id=$4) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_profile_schema_versions WHERE id=$5)`
		if err := database.QueryRowContext(
			ctx, query, adoptedOrganizationID(other), adoptedID(other, 2),
			adoptedID(other, 10), adoptedID(other, 14), adoptedID(other, 16),
		).Scan(&leaked); err != nil {
			t.Fatalf("read IAM V1 adoption isolation: %v", err)
		}
		if leaked != 0 {
			t.Fatalf("IAM V1 adoption leaked from %s into %s", other.schema, tenant.schema)
		}
	}
}

func assertIAMV1AdoptionPrivileges(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var runtimeTable, runtimeRoutine, adopterTable, adopterRoutine bool
	query := `SELECT
 has_table_privilege('harpy_runtime','t_iam_a.iam_v1_adoption_stage','INSERT'),
 has_function_privilege('harpy_runtime',
  't_iam_a.iam_v1_adoption_project_v1(name,uuid,uuid)','EXECUTE'),
 has_table_privilege('harpy_adopter','t_iam_a.iam_v1_adoption_stage','INSERT'),
 has_function_privilege('harpy_adopter',
  't_iam_a.iam_v1_adoption_project_v1(name,uuid,uuid)','EXECUTE')`
	if err := database.QueryRowContext(ctx, query).Scan(
		&runtimeTable, &runtimeRoutine, &adopterTable, &adopterRoutine,
	); err != nil {
		t.Fatalf("read IAM V1 adoption privileges: %v", err)
	}
	if runtimeTable || runtimeRoutine || adopterTable || !adopterRoutine {
		t.Fatalf("IAM V1 adoption privilege closure = %v/%v/%v/%v",
			runtimeTable, runtimeRoutine, adopterTable, adopterRoutine)
	}
}
