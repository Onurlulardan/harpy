//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func rollbackProjectAndVerifyIAMV1Adoption(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, operationID, runID string,
) string {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM V1 rollback rehearsal: %v", err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin IAM V1 rollback rehearsal: %v", err)
	}
	defer transaction.Rollback()
	bindIAMV1Adopter(t, ctx, transaction, tenant.schema)
	project := "SELECT " + tenant.schema +
		".iam_v1_adoption_project_v1($1::name,$2::uuid,$3::uuid)"
	var projected int
	if err := transaction.QueryRowContext(
		ctx, project, tenant.schema, operationID, runID,
	).Scan(&projected); err != nil || projected != adoptedTargetCount(tenant) {
		t.Fatalf("project IAM V1 rollback rehearsal = %d/%v", projected, err)
	}
	verify := "SELECT " + tenant.schema +
		".iam_v1_adoption_verify_v1($1::name,$2::uuid,$3::uuid)"
	var paritySHA string
	if err := transaction.QueryRowContext(
		ctx, verify, tenant.schema, operationID, runID,
	).Scan(&paritySHA); err != nil || len(paritySHA) != 64 {
		t.Fatalf("verify IAM V1 rollback rehearsal = %q/%v", paritySHA, err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatalf("rollback IAM V1 adoption rehearsal: %v", err)
	}
	return paritySHA
}

func assertIAMV1AdoptionRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, operationID, runID string,
) {
	t.Helper()
	query := `SELECT
 (SELECT state FROM ` + tenant.schema + `.iam_v1_adoption_runs
   WHERE operation_id=$1 AND run_id=$2),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_users WHERE id=$3) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_organizations WHERE id=$4) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_permissions WHERE id=$5) +
 (SELECT count(*) FROM ` + tenant.schema + `.iam_invitations WHERE id=$6),
 (SELECT state FROM platform.cutover_groups
   WHERE scope_kind='tenant' AND scope_id=$7 AND group_id='iam_system_resources'),
 (SELECT writer FROM platform.cutover_groups
   WHERE scope_kind='tenant' AND scope_id=$7 AND group_id='iam_system_resources')`
	var state, groupState, writer string
	var projected int
	if err := database.QueryRowContext(ctx, query, operationID, runID,
		adoptedID(tenant, 2), adoptedID(tenant, 1), adoptedID(tenant, 10),
		adoptedID(tenant, 14), tenant.id,
	).Scan(&state, &projected, &groupState, &writer); err != nil {
		t.Fatalf("read IAM V1 adoption rollback state: %v", err)
	}
	if state != "copying" || projected != 0 || groupState != "shadow" || writer != "v1" {
		t.Fatalf("IAM V1 rollback state = %s/%d/%s/%s",
			state, projected, groupState, writer)
	}
}
