//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/adoption"
)

var iamV1AdoptionCategories = []string{
	"organizations", "roles", "users", "memberships", "assignments",
	"identifiers", "portal", "credentials", "resources", "actions",
	"permissions", "permission_requests", "invitations", "profiles", "layouts",
}

func registerIAMV1AdoptionPlan(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	operationID string,
	runID string,
) {
	t.Helper()
	categories := make([]adoptionapp.CategoryEvidence, 0, len(iamV1AdoptionCategories))
	for _, category := range iamV1AdoptionCategories {
		count := len(adoptedTargets(tenant, category))
		if category == "layouts" {
			count = 1
		}
		categories = append(categories, adoptionapp.CategoryEvidence{
			Category: category, RowCount: count, RootSHA256: strings.Repeat("3", 64),
		})
	}
	plan := adoptionapp.Plan{
		ContractVersion: adoptionapp.PlanContractVersion,
		TenantID:        tenant.id, MappingSHA256: strings.Repeat("1", 64),
		SourceSHA256: strings.Repeat("2", 64), Categories: categories,
		Issues: make([]adoptionapp.Issue, 0), Eligible: true,
	}
	document, digest, err := adoptionapp.EncodePlan(plan)
	if err != nil {
		t.Fatalf("encode IAM V1 adoption plan: %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM V1 adoption planner: %v", err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin IAM V1 adoption planner: %v", err)
	}
	defer transaction.Rollback()
	bindIAMV1Adopter(t, ctx, transaction, tenant.schema)
	var state string
	query := "SELECT " + tenant.schema + ".iam_v1_adoption_plan_register_v1(" +
		"$1::name,$2::uuid,$3::uuid,$4::uuid,$5::bytea,$6::text)"
	if err := transaction.QueryRowContext(
		ctx, query, tenant.schema, operationID, runID, tenant.id, document, digest,
	).Scan(&state); err != nil {
		t.Fatalf("register IAM V1 adoption plan in %s: %v", tenant.schema, err)
	}
	if state != "planned" {
		t.Fatalf("IAM V1 adoption plan state = %s", state)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM V1 adoption plan: %v", err)
	}
}

func bindIAMV1Adopter(
	t *testing.T, ctx context.Context, transaction *sql.Tx, schema string,
) {
	t.Helper()
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		schema+", pg_catalog"); err != nil {
		t.Fatalf("bind IAM V1 adoption schema: %v", err)
	}
	if _, err := transaction.ExecContext(
		ctx, "SET LOCAL SESSION AUTHORIZATION harpy_adopter",
	); err != nil {
		t.Fatalf("bind IAM V1 adopter: %v", err)
	}
}
