//go:build integration

package postgres

import (
	"context"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func createSchemaExecutorIndexTable(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	planValue schemav1.Plan,
	duplicates bool,
) {
	t.Helper()
	mapping := planValue.Steps[0].IdentifierMapping
	table := pgIdentifier(schemaName) + "." + pgIdentifier(mapping.Form.PhysicalName)
	field := pgIdentifier(mapping.Fields[0].PhysicalName)
	statement := "CREATE TABLE " + table + ` (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    row_version bigint NOT NULL DEFAULT 1 CHECK (row_version > 0),
    deleted_at timestamptz,
    ` + field + ` text
)`
	if _, err := environment.pool.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
	value := "value::text"
	if duplicates {
		value = "'duplicate'::text"
	}
	insert := "INSERT INTO " + table + " (id, " + field + ") " +
		"SELECT pg_catalog.gen_random_uuid(), " + value +
		" FROM pg_catalog.generate_series(1, 4) AS value"
	if _, err := environment.pool.Exec(ctx, insert); err != nil {
		t.Fatal(err)
	}
}

func createSchemaExecutorExactIndex(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	planValue schemav1.Plan,
	expectFailure bool,
) {
	t.Helper()
	mapping := planValue.Steps[0].IdentifierMapping
	arguments := schemaExecutorIndexArgumentsFromPlan(t, planValue)
	unique := ""
	if arguments.Unique {
		unique = "UNIQUE "
	}
	statement := "CREATE " + unique + "INDEX CONCURRENTLY " +
		pgIdentifier(mapping.Indexes[0].PhysicalName) + " ON " +
		pgIdentifier(schemaName) + "." + pgIdentifier(mapping.Form.PhysicalName) +
		" USING btree (" + pgIdentifier(mapping.Fields[0].PhysicalName) +
		") WHERE deleted_at IS NULL"
	_, err := environment.pool.Exec(ctx, statement)
	if expectFailure && err == nil {
		t.Fatal("expected invalid unique concurrent index")
	}
	if !expectFailure && err != nil {
		t.Fatalf("precreate exact concurrent index: %v", err)
	}
}

func schemaExecutorIndexArgumentsFromPlan(
	t *testing.T,
	planValue schemav1.Plan,
) schemav1.IndexArguments {
	t.Helper()
	document, err := planValue.Steps[0].Arguments["unique"].MarshalJSON()
	if err != nil || len(document) == 0 {
		t.Fatalf("read index arguments: %v", err)
	}
	return schemav1.IndexArguments{
		Unique: string(document) == "true", ActiveRecordsOnly: true,
	}
}

func assertSchemaExecutorIndexValid(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
	planValue schemav1.Plan,
) {
	t.Helper()
	mapping := planValue.Steps[0].IdentifierMapping
	contract, exists, err := schemaExecutorReadIndex(
		ctx, environment.pool, schemaName, mapping.Indexes[0].PhysicalName,
	)
	if err != nil || !exists || !contract.valid || !contract.ready || !contract.standalone {
		t.Fatalf("index contract=%#v exists=%t err=%v", contract, exists, err)
	}
}
