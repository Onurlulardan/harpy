//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMV1AdoptedMatrix(
	t *testing.T, ctx context.Context, database *sql.DB, tenant iamFoundationTenant,
) {
	t.Helper()
	query := `SELECT
 (SELECT count(*) FROM ` + tenant.schema + `.iam_users WHERE id=$1),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_user_identifiers WHERE id=$2),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_password_credentials WHERE user_id=$1),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_organization_memberships WHERE id=$3),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_roles WHERE id=$4),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_user_role_assignments WHERE id=$5),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_permission_requests WHERE id=$6),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_invitations WHERE id=$7),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_profile_field_revisions
   WHERE schema_version_id=$8 AND field_id=$9),
 (SELECT count(*) FROM ` + tenant.schema + `.iam_user_permission_bindings AS binding
   JOIN ` + tenant.schema + `.iam_permissions AS permission
     ON permission.id=binding.permission_id AND permission.state='active'
   JOIN ` + tenant.schema + `.iam_resources AS resource
     ON resource.id=permission.resource_id AND resource.state='active'
   JOIN ` + tenant.schema + `.iam_actions AS action
     ON action.id=permission.action_id AND action.state='active'
   WHERE binding.user_id=$1 AND binding.state='active')`
	counts := make([]int, 10)
	if err := database.QueryRowContext(ctx, query,
		adoptedID(tenant, 2), adoptedID(tenant, 3), adoptedID(tenant, 4),
		adoptedID(tenant, 5), adoptedID(tenant, 6), adoptedID(tenant, 13),
		adoptedID(tenant, 14), adoptedID(tenant, 16), adoptedID(tenant, 17),
	).Scan(&counts[0], &counts[1], &counts[2], &counts[3], &counts[4],
		&counts[5], &counts[6], &counts[7], &counts[8], &counts[9]); err != nil {
		t.Fatalf("read IAM V1 adopted matrix in %s: %v", tenant.schema, err)
	}
	for index, count := range counts {
		if count != 1 {
			t.Fatalf("IAM V1 adopted matrix item %d count = %d", index, count)
		}
	}
	assertIAMV1AdoptionCategoryEvidence(t, ctx, database, tenant)
}

func assertIAMV1AdoptionCategoryEvidence(
	t *testing.T, ctx context.Context, database *sql.DB, tenant iamFoundationTenant,
) {
	t.Helper()
	var populated, copied int
	query := `SELECT
	count(*) FILTER (WHERE source_count > 0),
	count(*) FILTER (WHERE category <> 'layouts' AND copy_state='complete'
	  AND source_count=staged_count AND staged_count > 0 AND target_count=staged_count
   AND staged_sha256=target_sha256)
 FROM ` + tenant.schema + `.iam_v1_adoption_category_evidence`
	if err := database.QueryRowContext(ctx, query).Scan(&populated, &copied); err != nil {
		t.Fatalf("read IAM V1 adoption evidence: %v", err)
	}
	if populated != len(iamV1AdoptionCategories) || copied != len(iamV1AdoptionCategories)-1 {
		t.Fatalf("IAM V1 adoption evidence = populated %d copied %d", populated, copied)
	}
}
