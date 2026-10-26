//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func prepareProtectedAuthorityProbe(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	_, err := admin.Exec(ctx, `CREATE SCHEMA cutover_authority
AUTHORIZATION harpy_cutover_authority_owner;
CREATE TABLE cutover_authority.authority_probe (id integer PRIMARY KEY);
ALTER TABLE cutover_authority.authority_probe OWNER TO harpy_cutover_authority_owner;
INSERT INTO cutover_authority.authority_probe VALUES (1);
CREATE FUNCTION cutover_authority.authority_probe_v1() RETURNS integer
LANGUAGE sql SET search_path=pg_catalog AS 'SELECT 1';
ALTER FUNCTION cutover_authority.authority_probe_v1()
OWNER TO harpy_cutover_authority_owner;
GRANT USAGE ON SCHEMA cutover_authority TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION cutover_authority.authority_probe_v1()
TO harpy_cluster_admin;`)
	if err != nil {
		t.Fatalf("prepare protected authority probe: %v", err)
	}
}

func assertAuthorityProbeExecuteGrant(
	t *testing.T, ctx context.Context, clusterAdmin *pgx.Conn,
) {
	t.Helper()
	var result int
	if err := clusterAdmin.QueryRow(ctx,
		"SELECT cutover_authority.authority_probe_v1()").Scan(&result); err != nil || result != 1 {
		t.Fatalf("cluster admin explicit execute grant result=%d err=%v", result, err)
	}
}

func assertClusterAdminCannotMutateAuthority(
	t *testing.T, ctx context.Context, clusterAdmin *pgx.Conn,
) {
	t.Helper()
	statements := []string{
		"SET ROLE harpy_cutover_authority_owner",
		"GRANT harpy_cutover_authority_owner TO harpy_owner",
		"ALTER ROLE harpy_cutover_authority_owner LOGIN",
		"ALTER ROLE harpy_cutover_authority_owner PASSWORD 'takeover'",
		"ALTER ROLE harpy_cutover_authority_owner CREATEROLE",
		"ALTER ROLE harpy_cutover_authority_owner RENAME TO stolen_authority_owner",
		"COMMENT ON ROLE harpy_cutover_authority_owner IS 'mutated'",
		"DROP ROLE harpy_cutover_authority_owner",
		"REASSIGN OWNED BY harpy_cutover_authority_owner TO harpy_cluster_admin",
		"DROP OWNED BY harpy_cutover_authority_owner",
		"DROP SCHEMA cutover_authority CASCADE",
		"DROP TABLE cutover_authority.authority_probe",
		"DROP FUNCTION cutover_authority.authority_probe_v1()",
		"ALTER SCHEMA cutover_authority OWNER TO harpy_cluster_admin",
		"ALTER TABLE cutover_authority.authority_probe OWNER TO harpy_cluster_admin",
		"ALTER FUNCTION cutover_authority.authority_probe_v1() OWNER TO harpy_cluster_admin",
		`CREATE OR REPLACE FUNCTION cutover_authority.authority_probe_v1()
RETURNS integer LANGUAGE sql SET search_path=pg_catalog AS 'SELECT 2'`,
		"INSERT INTO cutover_authority.authority_probe VALUES (2)",
		"UPDATE cutover_authority.authority_probe SET id=2 WHERE id=1",
		"DELETE FROM cutover_authority.authority_probe WHERE id=1",
		"TRUNCATE cutover_authority.authority_probe",
		`ALTER DEFAULT PRIVILEGES FOR ROLE harpy_cutover_authority_owner
IN SCHEMA cutover_authority GRANT SELECT ON TABLES TO harpy_owner`,
	}
	for _, statement := range statements {
		assertClusterAdminDeniedSQL(t, ctx, clusterAdmin, statement)
	}
}

func assertClusterAdminCannotChangeAuthorityACL(
	t *testing.T, ctx context.Context, admin, clusterAdmin *pgx.Conn,
) {
	t.Helper()
	before := authorityProbeACL(t, ctx, admin)
	statements := []string{
		"GRANT CREATE ON SCHEMA cutover_authority TO harpy_owner",
		"REVOKE USAGE ON SCHEMA cutover_authority FROM harpy_cluster_admin",
		"GRANT SELECT ON cutover_authority.authority_probe TO harpy_owner",
		"REVOKE ALL ON cutover_authority.authority_probe FROM harpy_owner",
		"GRANT EXECUTE ON FUNCTION cutover_authority.authority_probe_v1() TO harpy_owner",
		"REVOKE EXECUTE ON FUNCTION cutover_authority.authority_probe_v1() FROM harpy_cluster_admin",
	}
	for _, statement := range statements {
		_, err := clusterAdmin.Exec(ctx, statement)
		if err == nil {
			continue
		}
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
			t.Fatalf("cluster admin ACL mutation %q error=%v", statement, err)
		}
	}
	if after := authorityProbeACL(t, ctx, admin); after != before {
		t.Fatalf("authority ACL changed before=%q after=%q", before, after)
	}
}

func authorityProbeACL(t *testing.T, ctx context.Context, admin *pgx.Conn) string {
	t.Helper()
	var acl string
	err := admin.QueryRow(ctx, `SELECT concat_ws('|', schema.nspacl::text,
relation.relacl::text, procedure.proacl::text)
FROM pg_catalog.pg_namespace AS schema
JOIN pg_catalog.pg_class AS relation ON relation.relnamespace=schema.oid
JOIN pg_catalog.pg_proc AS procedure ON procedure.pronamespace=schema.oid
WHERE schema.nspname='cutover_authority' AND relation.relname='authority_probe'
AND procedure.proname='authority_probe_v1'`).Scan(&acl)
	if err != nil {
		t.Fatal(err)
	}
	return acl
}

func assertClusterAdminDeniedSQL(
	t *testing.T, ctx context.Context, clusterAdmin *pgx.Conn, statement string,
) {
	t.Helper()
	_, err := clusterAdmin.Exec(ctx, statement)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("cluster admin mutation %q error=%v", statement, err)
	}
}
