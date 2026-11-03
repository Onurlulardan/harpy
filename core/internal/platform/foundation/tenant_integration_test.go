//go:build integration

package foundation

import (
	"strings"
	"testing"
)

func TestTenantNamespacesAreIsolatedAndStable(t *testing.T) {
	_, ctx := adminConn(t)
	cfg := testConfig(databaseName(t))
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	leftID := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	rightID := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	left, err := PrepareTenantNamespace(ctx, cfg, leftID, "tenant-left-foundation-secret")
	if err != nil {
		t.Fatal(err)
	}
	again, err := PrepareTenantNamespace(ctx, cfg, leftID, "tenant-left-foundation-secret")
	if err != nil || again != left {
		t.Fatalf("namespace changed: %+v %+v %v", left, again, err)
	}
	right, err := PrepareTenantNamespace(ctx, cfg, rightID, "tenant-right-foundation-secret")
	if err != nil || right.Schema == left.Schema {
		t.Fatal(err)
	}
	if err := Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	leftConn := connectRole(t, ctx, left.Role, "tenant-left-foundation-secret", cfg.Database)
	if _, err := leftConn.Exec(ctx, "CREATE TABLE "+ident(left.Schema)+".item(id integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := leftConn.Exec(ctx, "INSERT INTO "+ident(left.Schema)+".item(id) VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	rightConn := connectRole(t, ctx, right.Role, "tenant-right-foundation-secret", cfg.Database)
	if _, err := rightConn.Exec(ctx, "SELECT id FROM "+ident(left.Schema)+".item"); err == nil {
		t.Fatal("tenant read another tenant schema")
	}
	if _, err := rightConn.Exec(ctx, "CREATE TABLE "+ident(left.Schema)+".stolen(id integer)"); err == nil {
		t.Fatal("tenant created a table in another schema")
	}
	if _, err := leftConn.Exec(ctx, "CREATE SCHEMA injected_name"); err == nil {
		t.Fatal("tenant created a schema")
	}
	api := connectRole(t, ctx, RoleAPI, apiPassword, cfg.Database)
	var id int
	if err := api.QueryRow(ctx, "SELECT id FROM "+ident(left.Schema)+".item").Scan(&id); err != nil || id != 1 {
		t.Fatal(err)
	}
	if _, err := api.Exec(ctx, "CREATE TABLE "+ident(left.Schema)+".blocked(id integer)"); err == nil {
		t.Fatal("api created a tenant table")
	}
	goose := connectRole(t, ctx, RoleGoose, goosePassword, cfg.Database)
	if _, err := goose.Exec(ctx, "CREATE TABLE "+ident(right.Schema)+".blocked(id integer)"); err == nil {
		t.Fatal("goose created a tenant table")
	}
	schema, err := TenantSchemaName(leftID)
	if err != nil {
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
	if _, err := target.Exec(ctx, "DROP SCHEMA "+ident(schema)+" CASCADE"); err != nil {
		t.Fatal(err)
	}
	if _, err := target.Exec(ctx, "CREATE SCHEMA "+ident(schema)); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareTenantNamespace(ctx, cfg, leftID, "tenant-left-foundation-secret"); err == nil ||
		!strings.Contains(err.Error(), "owned by") {
		t.Fatal(err)
	}
}
