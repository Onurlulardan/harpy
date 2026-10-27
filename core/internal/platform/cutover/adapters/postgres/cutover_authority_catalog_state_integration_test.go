//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

type builtAuthorityQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type builtAuthorityCatalogState struct {
	identity   builtAuthorityIdentity
	oid        uint32
	owner      string
	definition string
	acl        string
}

func readBuiltAuthorityCatalogStates(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
	identities []builtAuthorityIdentity,
) []builtAuthorityCatalogState {
	t.Helper()
	states := make([]builtAuthorityCatalogState, len(identities))
	for index, identity := range identities {
		states[index] = readBuiltAuthorityCatalogState(t, ctx, query, identity)
	}
	return states
}

func readBuiltAuthorityCatalogState(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
	identity builtAuthorityIdentity,
) builtAuthorityCatalogState {
	t.Helper()
	statement := builtAuthorityLocateStatement(t, identity.kind)
	state := builtAuthorityCatalogState{identity: identity}
	err := query.QueryRow(ctx, statement,
		identity.schema, identity.name, identity.arguments,
	).Scan(&state.oid, &state.owner)
	if err != nil {
		var candidates string
		_ = query.QueryRow(ctx, `SELECT COALESCE(pg_catalog.string_agg(
    pg_catalog.oidvectortypes(procedure_row.proargtypes), E'\n' ORDER BY procedure_row.oid), '')
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname=$1 AND procedure_row.proname=$2`,
			identity.schema, identity.name,
		).Scan(&candidates)
		t.Fatalf("locate authority %s %s.%s(%s): %v candidates=%q", identity.kind,
			identity.schema, identity.name, identity.arguments, err, candidates)
	}
	err = query.QueryRow(ctx, `SELECT
    cutover_authority.catalog_definition_sha256_v1($1::text,$2::oid),
    cutover_authority.catalog_acl_sha256_v1($1::text,$2::oid)`,
		identity.kind, state.oid,
	).Scan(&state.definition, &state.acl)
	if err != nil {
		t.Fatalf("fingerprint authority %s %s.%s(%s): %v", identity.kind,
			identity.schema, identity.name, identity.arguments, err)
	}
	return state
}

func builtAuthorityLocateStatement(t *testing.T, kind string) string {
	t.Helper()
	switch kind {
	case "schema":
		return `SELECT namespace.oid, pg_catalog.pg_get_userbyid(namespace.nspowner)
FROM pg_catalog.pg_namespace AS namespace
WHERE namespace.nspname=$1 AND $2=$1 AND $3=''`
	case "table":
		return `SELECT relation.oid, pg_catalog.pg_get_userbyid(relation.relowner)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname=$1 AND relation.relname=$2
  AND relation.relkind IN ('r','p') AND $3=''`
	case "function":
		return `SELECT procedure_row.oid, pg_catalog.pg_get_userbyid(procedure_row.proowner)
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname=$1 AND procedure_row.proname=$2
  AND pg_catalog.replace(pg_catalog.oidvectortypes(
      procedure_row.proargtypes),' ','')=$3`
	default:
		t.Fatalf("unsupported authority identity kind %s", kind)
		return ""
	}
}

func assertBuiltAuthorityStatesEqual(
	t *testing.T, label string,
	expected, actual []builtAuthorityCatalogState,
) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("%s authority state count = %d, want %d", label, len(actual), len(expected))
	}
	for index := range expected {
		if expected[index] != actual[index] {
			t.Fatalf("%s authority state %d differs: got=%#v want=%#v",
				label, index, actual[index], expected[index])
		}
	}
}
