//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func assertIAMProfileRevisionReads(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	historical := mustReadIAMProfileSchema(t, ctx, database, "t_iam_a", `
SELECT iam_profile_historical_read_v1(
    pg_catalog.current_schema()::name, $1
)::text`, iamUserProfileSchemaID)
	assertIAMProfileSchemaState(t, historical, 1,
		profileapp.SchemaRetired, profileapp.FieldActive, 64)

	active := mustReadIAMProfileSchema(t, ctx, database, "t_iam_a", `
SELECT iam_profile_active_catalog_read_v1(
    pg_catalog.current_schema()::name, $1
)::text`, "user")
	assertIAMProfileSchemaState(t, active, 2,
		profileapp.SchemaActive, profileapp.FieldRetired, 96)

	tenantB := mustReadIAMProfileSchema(t, ctx, database, "t_iam_b", `
SELECT iam_profile_active_catalog_read_v1(
    pg_catalog.current_schema()::name, $1
)::text`, "user")
	assertIAMProfileSchemaState(t, tenantB, 1,
		profileapp.SchemaActive, profileapp.FieldActive, 128)
	assertIAMHistoricalProfileDocument(t, ctx, database, historical)
}

func mustReadIAMProfileSchema(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	argument any,
) profileapp.Schema {
	t.Helper()
	payload := mustRunIAMAtomicText(t, ctx, database, schema, statement, argument)
	decoded, err := decodeProfileSchemaSnapshot([]byte(payload))
	if err != nil {
		t.Fatalf("decode %s IAM profile schema: %v", schema, err)
	}
	return decoded
}

func assertIAMProfileSchemaState(
	t *testing.T,
	schema profileapp.Schema,
	version int64,
	schemaState profileapp.SchemaState,
	fieldState profileapp.FieldState,
	maximumBytes int,
) {
	t.Helper()
	fields := schema.Fields()
	if len(fields) != 1 {
		t.Fatalf("profile schema field count = %d, want 1", len(fields))
	}
	wantValidation := `{"maximum_bytes":` + strconv.Itoa(maximumBytes) + `}`
	if schema.Version() != version || schema.State() != schemaState ||
		fields[0].State() != fieldState ||
		string(fields[0].ValidationJSON()) != wantValidation {
		t.Fatalf("profile schema v%d/%s fields=%d state=%v validation=%s",
			schema.Version(), schema.State(), len(fields),
			fields[0].State(), fields[0].ValidationJSON())
	}
}

func assertIAMHistoricalProfileDocument(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema profileapp.Schema,
) {
	t.Helper()
	values := `{"43000000-0000-4000-8000-000000000010":"Ada"}`
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
UPDATE iam_users SET profile_values = $1::jsonb,
    profile_sha256 = pg_catalog.sha256(pg_catalog.convert_to($2::text, 'UTF8'))
WHERE id = $3`, values, values, iamUserOneID)
	payload := mustRunIAMAtomicText(t, ctx, database, "t_iam_a", `
SELECT iam_profile_document_read_v1(
    pg_catalog.current_schema()::name, $1, $2
)::text`, "user", iamUserOneID)
	wire, schemaID, digest, err := decodeProfileDocument([]byte(payload))
	if err != nil || schemaID != schema.ID() {
		t.Fatalf("decode historical IAM document: schema=%s err=%v", schemaID, err)
	}
	document, err := profileapp.DecodeDocument(schema, wire.Values, digest, nil)
	if err != nil {
		unchecked, uncheckedErr := profileapp.DecodeDocument(schema, wire.Values, nil, nil)
		t.Fatalf("historical IAM document invalid: values=%s stored=%x computed=%x unchecked=%v",
			wire.Values, digest, unchecked.SHA256(), uncheckedErr)
	}
	value, found := document.Value(uuid.MustParse(iamProfileFieldID))
	if !found || string(value) != `"Ada"` {
		t.Fatalf("historical IAM document value=%s found=%v", value, found)
	}
}
