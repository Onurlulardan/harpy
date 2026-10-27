//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMProfileRetireRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamProfileCommandFixture,
	marker byte,
) {
	t.Helper()
	operationID := profileCommandOperationID(marker, 4)
	requestSHA256 := [32]byte{marker, 4}
	referenceSHA256 := [32]byte{marker, 0x7f}
	result, err := runIAMProfileRetire(
		ctx, database, tenant.schema, fixture.retired, fixture.published.ID(), 1,
		operationID, requestSHA256, referenceSHA256,
	)
	if err != nil || result.id != fixture.retired.ID() || result.revision != 3 ||
		result.clock != 3 || result.replayed {
		t.Fatalf("profile retirement %s = %+v, %v", tenant.schema, result, err)
	}
	replay, err := runIAMProfileRetire(
		ctx, database, tenant.schema, fixture.retired, fixture.published.ID(), 1,
		operationID, requestSHA256, referenceSHA256,
	)
	if err != nil || !replay.replayed || replay.id != result.id ||
		replay.revision != result.revision || replay.clock != result.clock {
		t.Fatalf("profile retirement replay %s = %+v, %v", tenant.schema, replay, err)
	}
	var schemaID, fieldState string
	var evidenceCount int
	mustQueryIAMTenantValue(t, ctx, database, tenant.schema, `
SELECT schema_version.id::text, field.state,
       (SELECT count(*) FROM iam_profile_retirement_evidence)
FROM iam_profile_schema_versions AS schema_version
JOIN iam_profile_fields AS field ON field.id='`+iamProfileCommandFieldID+`'
WHERE schema_version.resource_kind='user' AND schema_version.state='active'`,
		&schemaID, &fieldState, &evidenceCount)
	if schemaID != iamProfileCommandSchemaV3ID || fieldState != "retired" ||
		evidenceCount != 1 {
		t.Fatalf("profile retirement %s state = %s/%s/%d",
			tenant.schema, schemaID, fieldState, evidenceCount)
	}
}

func assertIAMProfileRetireTenantIsolation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	foreign iamFoundationTenant,
) {
	t.Helper()
	var schemaID, fieldState string
	mustQueryIAMTenantValue(t, ctx, database, foreign.schema, `
SELECT schema_version.id::text, field.state
FROM iam_profile_schema_versions AS schema_version
JOIN iam_profile_fields AS field ON field.id='`+iamProfileCommandFieldID+`'
WHERE schema_version.resource_kind='user' AND schema_version.state='active'`,
		&schemaID, &fieldState)
	if schemaID != iamProfileCommandSchemaV2ID || fieldState != "active" {
		t.Fatalf("foreign profile retirement state = %s/%s", schemaID, fieldState)
	}
}

func assertIAMProfileDirectDMLDenied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, executeErr := transaction.ExecContext(ctx, `
UPDATE iam_profile_fields SET revision=revision+1
WHERE id='`+iamProfileCommandFieldID+`'`)
	_ = transaction.Rollback()
	_ = connection.Close()
	if executeErr == nil {
		t.Fatal("IAM runtime received direct profile table DML")
	}
	assertIAMSQLState(t, executeErr, "42501")
}
