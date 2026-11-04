//go:build integration

package foundation

import (
	"strings"
	"testing"
)

func TestRuntimeCannotDDLAndGooseCanThroughOwner(t *testing.T) {
	_, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	goose := connectRole(t, ctx, RoleGoose, goosePassword, cfg.Database)
	if _, err := goose.Exec(ctx, "CREATE TABLE platform.direct(id integer)"); err == nil {
		t.Fatal("goose created a table without set role")
	}
	if _, err := goose.Exec(ctx, "SET ROLE "+ident(RoleOwner)); err != nil {
		t.Fatal(err)
	}
	if _, err := goose.Exec(ctx, "CREATE TABLE platform.probe(id integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := goose.Exec(ctx, "DROP TABLE platform.probe"); err != nil {
		t.Fatal(err)
	}
	for _, role := range []struct{ name, password string }{
		{RoleAPI, apiPassword}, {RoleWorker, workerPassword},
	} {
		runtime := connectRole(t, ctx, role.name, role.password, cfg.Database)
		for _, statement := range []string{
			"CREATE TABLE platform.denied(id integer)",
			"CREATE SCHEMA denied_schema",
			"DROP SCHEMA platform",
			"ALTER ROLE " + ident(role.name) + " SUPERUSER",
		} {
			if _, err := runtime.Exec(ctx, statement); err == nil {
				t.Fatalf("%s executed %s", role.name, statement)
			}
		}
	}
}

func TestApplyRefusesSuperuserAndExtraCreate(t *testing.T) {
	conn, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if _, err := conn.Exec(ctx, `CREATE ROLE harpy_api LOGIN SUPERUSER PASSWORD 'ignored-value'`); err != nil {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err == nil || !strings.Contains(err.Error(), "unexpected attributes") {
		t.Fatal(err)
	}
	var super bool
	if err := conn.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = 'harpy_api'`).Scan(&super); err != nil || !super {
		t.Fatalf("superuser role changed: %v %t", err, super)
	}
	if err := resetCluster(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	targetDSN, err := databaseDSN(adminDSN, cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	target, err := connect(ctx, targetDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close(ctx)
	if _, err := target.Exec(ctx, "GRANT CREATE ON SCHEMA platform TO "+ident(RoleAPI)); err != nil {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err == nil || !strings.Contains(err.Error(), "can create") {
		t.Fatal(err)
	}
	var create bool
	if err := target.QueryRow(ctx, `SELECT has_schema_privilege($1, 'platform', 'CREATE')`, RoleAPI).Scan(&create); err != nil || !create {
		t.Fatalf("extra privilege was rewritten: %v %t", err, create)
	}
}
