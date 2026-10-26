//go:build integration

package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type builtAuthorityPreview struct {
	preObjects   []builtAuthorityCatalogState
	postObjects  []builtAuthorityCatalogState
	dependencies []builtAuthorityCatalogState
	quiescence   builtAuthorityQuiescence
}

func previewBuiltAuthorityAdoption(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	quiescence builtAuthorityQuiescence,
) builtAuthorityPreview {
	t.Helper()
	transaction, err := database.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	lockBuiltAuthorityAdoption(t, ctx, transaction)
	assertBuiltAuthorityNamespaceCount(t, ctx, transaction)
	pre := readBuiltAuthorityCatalogStates(
		t, ctx, transaction, builtAuthorityObjectIdentities,
	)
	assertBuiltAuthorityOwners(t, "preview pre", pre, "harpy_owner")
	dependencies := readBuiltAuthorityCatalogStates(
		t, ctx, transaction, builtAuthorityDependencyIdentities,
	)
	assertBuiltAuthorityOwners(t, "preview dependency", dependencies, "harpy_owner")
	transferBuiltAuthorityObjects(t, ctx, transaction)
	post := readBuiltAuthorityCatalogStates(
		t, ctx, transaction, builtAuthorityObjectIdentities,
	)
	assertBuiltAuthorityOwners(t, "preview post", post, cutoverAuthorityOwner)
	afterDependencies := readBuiltAuthorityCatalogStates(
		t, ctx, transaction, builtAuthorityDependencyIdentities,
	)
	assertBuiltAuthorityStatesEqual(t, "preview dependency", dependencies, afterDependencies)
	return builtAuthorityPreview{
		preObjects: pre, postObjects: post, dependencies: dependencies,
		quiescence: quiescence,
	}
}

func lockBuiltAuthorityAdoption(
	t *testing.T, ctx context.Context, transaction pgx.Tx,
) {
	t.Helper()
	if _, err := transaction.Exec(ctx, `SELECT pg_catalog.pg_advisory_xact_lock(
    pg_catalog.hashtextextended('harpy.cutover_authority.adoption',0));
LOCK TABLE ONLY cutover_authority.authority_adoption_inventory,
    cutover_authority.finalization_authorizations,
    cutover_authority.finalization_receipts IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatalf("lock authority adoption catalog: %v", err)
	}
	tables := make([]string, 0, len(builtAuthorityDependencyIdentities))
	for _, identity := range builtAuthorityDependencyIdentities {
		if identity.kind == "table" {
			tables = append(tables, pgx.Identifier{identity.schema, identity.name}.Sanitize())
		}
	}
	statement := "LOCK TABLE ONLY " + strings.Join(tables, ",") + " IN SHARE MODE"
	if _, err := transaction.Exec(ctx, statement); err != nil {
		t.Fatalf("lock authority dependencies: %v", err)
	}
}

func transferBuiltAuthorityObjects(
	t *testing.T, ctx context.Context, transaction pgx.Tx,
) {
	t.Helper()
	for _, identity := range builtAuthorityObjectIdentities {
		qualified := pgx.Identifier{identity.schema, identity.name}.Sanitize()
		var statement string
		switch identity.kind {
		case "schema":
			statement = fmt.Sprintf("ALTER SCHEMA %s OWNER TO %s",
				pgx.Identifier{identity.schema}.Sanitize(), cutoverAuthorityOwner)
		case "table":
			statement = fmt.Sprintf("ALTER TABLE %s OWNER TO %s", qualified, cutoverAuthorityOwner)
		case "function":
			var functionIdentity string
			err := transaction.QueryRow(ctx, `SELECT procedure_row.oid::regprocedure::text
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname=$1 AND procedure_row.proname=$2
  AND pg_catalog.replace(pg_catalog.oidvectortypes(
      procedure_row.proargtypes),' ','')=$3`,
				identity.schema, identity.name, identity.arguments,
			).Scan(&functionIdentity)
			if err != nil {
				t.Fatalf("resolve authority function %s: %v", qualified, err)
			}
			statement = fmt.Sprintf("ALTER FUNCTION %s OWNER TO %s",
				functionIdentity, cutoverAuthorityOwner)
		default:
			t.Fatalf("unsupported authority transfer kind %s", identity.kind)
		}
		if _, err := transaction.Exec(ctx, statement); err != nil {
			t.Fatalf("transfer authority %s %s: %v", identity.kind, qualified, err)
		}
	}
}

func assertBuiltAuthorityNamespaceCount(
	t *testing.T, ctx context.Context, query builtAuthorityQuerier,
) {
	t.Helper()
	var count int
	err := query.QueryRow(ctx, `SELECT 1 +
    (SELECT count(*) FROM pg_catalog.pg_class AS relation
     JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
     WHERE namespace.nspname='cutover_authority' AND relation.relkind IN ('r','p')) +
    (SELECT count(*) FROM pg_catalog.pg_proc AS procedure_row
     JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
	    WHERE namespace.nspname='cutover_authority')`).Scan(&count)
	if err != nil || count != len(builtAuthorityObjectIdentities) {
		var catalog string
		_ = query.QueryRow(ctx, `SELECT pg_catalog.string_agg(identity, E'\n' ORDER BY identity)
FROM (
    SELECT 'schema:' || namespace.nspname AS identity
    FROM pg_catalog.pg_namespace AS namespace
    WHERE namespace.nspname='cutover_authority'
    UNION ALL
    SELECT 'table:' || relation.relname
    FROM pg_catalog.pg_class AS relation
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
    WHERE namespace.nspname='cutover_authority' AND relation.relkind IN ('r','p')
    UNION ALL
    SELECT 'function:' || procedure_row.proname || '(' ||
           pg_catalog.oidvectortypes(procedure_row.proargtypes) || ')'
    FROM pg_catalog.pg_proc AS procedure_row
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
    WHERE namespace.nspname='cutover_authority'
) AS actual`).Scan(&catalog)
		t.Fatalf("authority namespace count=%d want=%d err=%v actual=%s",
			count, len(builtAuthorityObjectIdentities), err, catalog)
	}
}

func assertBuiltAuthorityOwners(
	t *testing.T, label string, states []builtAuthorityCatalogState, expected string,
) {
	t.Helper()
	for index, state := range states {
		if state.owner != expected {
			t.Fatalf("%s owner %d=%s want=%s", label, index, state.owner, expected)
		}
	}
}
