//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func createSchemaExecutorBackfillTable(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	planValue schemav1.Plan,
	rows int,
) {
	t.Helper()
	mapping := planValue.Steps[0].IdentifierMapping
	table := pgIdentifier(schemaName) + "." + pgIdentifier(mapping.Form.PhysicalName)
	source, target := pgIdentifier(mapping.Fields[0].PhysicalName),
		pgIdentifier(mapping.Fields[1].PhysicalName)
	statement := "CREATE TABLE " + table + ` (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    row_version bigint NOT NULL DEFAULT 1 CHECK (row_version > 0),
    deleted_at timestamptz,
    ` + source + ` bigint,
    ` + target + ` bigint
)`
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatalf("create backfill table: %v", err)
	}
	insert := "INSERT INTO " + table + " (id, " + source + ") " +
		"SELECT pg_catalog.gen_random_uuid(), value " +
		"FROM pg_catalog.generate_series(1, $1) AS value"
	if _, err := environment.pool.Exec(ctx, insert, rows); err != nil {
		t.Fatalf("seed backfill table: %v", err)
	}
}

func installSchemaExecutorSuppressionTrigger(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	tableName string,
) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
CREATE FUNCTION platform.schema_executor_test_suppress_update()
RETURNS trigger LANGUAGE plpgsql SET search_path TO pg_catalog AS $function$
BEGIN
    RETURN NULL;
END;
$function$`)
	if err != nil {
		t.Fatal(err)
	}
	statement := "CREATE TRIGGER schema_executor_test_suppress_update " +
		"BEFORE UPDATE ON " + pgIdentifier(schemaName) + "." + pgIdentifier(tableName) +
		" FOR EACH ROW EXECUTE FUNCTION platform.schema_executor_test_suppress_update()"
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
}
