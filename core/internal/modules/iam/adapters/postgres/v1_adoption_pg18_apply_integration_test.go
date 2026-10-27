//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

var iamV1AdoptionMigrationNames = []string{
	"20260716013900_create_iam_v1_adoption_evidence",
	"20260716013901_create_iam_v1_adoption_stage",
	"20260716013902_create_iam_v1_adoption_plan_register_v1",
	"20260716013903_create_iam_v1_adoption_target_record_valid_v1",
	"20260716013904_create_iam_v1_adoption_batch_decode_v1",
	"20260716013905_create_iam_v1_adoption_batch_assert_v1",
	"20260716013906_create_iam_v1_adoption_apply_v1",
	"20260716013907_create_iam_v1_adoption_composite_v1",
	"20260716013908_create_iam_v1_adoption_project_identity_v1",
	"20260716013909_create_iam_v1_adoption_project_access_v1",
	"20260716013910_create_iam_v1_adoption_project_v1",
	"20260716013911_create_iam_v1_adoption_identity_match_v1",
	"20260716013912_create_iam_v1_adoption_access_match_v1",
	"20260716013913_create_iam_v1_adoption_verify_v1",
	"20260716013914_grant_iam_v1_adoption_v1",
}

func applyIAMV1Adoption(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamV1AdoptionMigrationNames))
	for _, name := range iamV1AdoptionMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
