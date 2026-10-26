//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func projectAndVerifyIAMV1Adoption(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, operationID, runID string,
) string {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM V1 adoption projector: %v", err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin IAM V1 adoption projection: %v", err)
	}
	defer transaction.Rollback()
	bindIAMV1Adopter(t, ctx, transaction, tenant.schema)
	var projected int64
	project := "SELECT " + tenant.schema +
		".iam_v1_adoption_project_v1($1::name,$2::uuid,$3::uuid)"
	if err := transaction.QueryRowContext(
		ctx, project, tenant.schema, operationID, runID,
	).Scan(&projected); err != nil {
		t.Fatalf("project IAM V1 adoption in %s: %v", tenant.schema, err)
	}
	wantProjected := int64(adoptedTargetCount(tenant))
	if projected != wantProjected {
		t.Fatalf("IAM V1 adoption projected = %d, want %d", projected, wantProjected)
	}
	var paritySHA string
	verify := "SELECT " + tenant.schema +
		".iam_v1_adoption_verify_v1($1::name,$2::uuid,$3::uuid)"
	if err := transaction.QueryRowContext(
		ctx, verify, tenant.schema, operationID, runID,
	).Scan(&paritySHA); err != nil {
		t.Fatalf("verify IAM V1 adoption in %s: %v", tenant.schema, err)
	}
	if len(paritySHA) != 64 {
		t.Fatalf("IAM V1 adoption parity digest = %q", paritySHA)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM V1 adoption projection: %v", err)
	}
	return paritySHA
}
