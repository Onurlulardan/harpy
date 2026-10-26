//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMProfileRevisionBackfill(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		var key, state, validation string
		var semanticValid, validationValid, definitionValid bool
		mustQueryIAMTenantValue(t, ctx, database, tenant.schema, `
SELECT snapshot.field_key, snapshot.field_state,
       snapshot.validation_schema::text,
       snapshot.semantic_sha256 = iam_profile_semantic_sha256_v1(
           snapshot.field_id, snapshot.resource_kind, snapshot.field_key,
           snapshot.value_kind, snapshot.cardinality, snapshot.required,
           snapshot.classification, snapshot.normalization_profile
       ),
       snapshot.validation_sha256 = pg_catalog.sha256(pg_catalog.convert_to(
           iam_profile_validation_canonical_v1(snapshot.validation_schema),
           'UTF8'
       )),
       schema_version.definition_sha256 =
           iam_profile_definition_sha256_v1(schema_version.id)
FROM iam_profile_field_revisions AS snapshot
JOIN iam_profile_schema_versions AS schema_version
  ON schema_version.id = snapshot.schema_version_id
WHERE snapshot.schema_version_id =
      '41000000-0000-4000-8000-000000000001'
  AND snapshot.field_id = '43000000-0000-4000-8000-000000000010'`,
			&key, &state, &validation,
			&semanticValid, &validationValid, &definitionValid)
		wantValidation := []string{
			`{"maximum_bytes": 64}`,
			`{"maximum_bytes": 128}`,
		}[index]
		if key != "display_name" || state != "active" ||
			validation != wantValidation || !semanticValid ||
			!validationValid || !definitionValid {
			t.Fatalf("%s profile backfill = %q/%q/%q/%v/%v/%v",
				tenant.schema, key, state, validation,
				semanticValid, validationValid, definitionValid)
		}
	}
}
