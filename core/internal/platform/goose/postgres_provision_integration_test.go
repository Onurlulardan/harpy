//go:build integration

package goose

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	lifecycle "github.com/Onurlulardan/harpy/core/internal/platform/initialize"
	"github.com/jackc/pgx/v5"
)

func TestInitializeRunsGooseOnEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_initialize_lifecycle")
	first := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	second := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	foundationPasses := 0
	request := lifecycle.Request{
		Config: cfg,
		Foundation: func(ctx context.Context, cfg foundation.Config) error {
			if err := foundation.Apply(ctx, cfg); err != nil {
				return err
			}
			foundationPasses++
			if foundationPasses == 1 {
				assertGooseHistoryCount(t, ctx, cfg, 0)
			}
			return nil
		},
		Goose:       func(ctx context.Context) error { return UpPlatform(ctx, cfg) },
		Application: lifecycle.PermitApplication,
	}
	if _, err := lifecycle.Run(ctx, request); err != nil {
		t.Fatal(err)
	}
	if _, err := lifecycle.Run(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := ProvisionTenant(ctx, cfg, first); err != nil {
		t.Fatal(err)
	}
	if err := ProvisionTenant(ctx, cfg, second); err != nil {
		t.Fatal(err)
	}
	assertTenantHistories(t, ctx, cfg, first, second)
	assertRuntimeDDLDenied(t, ctx, cfg)
}

func TestTenantAdvisoryLockBlocksConcurrentProvision(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_provision_lock")
	if err := foundation.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := UpPlatform(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	tenant := [16]byte{9, 9, 9, 9, 8, 8, 8, 8, 7, 7, 7, 7, 6, 6, 6, 6}
	conn := gooseAdmin(t, cfg)
	keyA, keyB := tenantAdvisoryKeys(tenant)
	var locked bool
	err := conn.QueryRow(ctx, "SELECT pg_catalog.pg_try_advisory_lock($1, $2)", keyA, keyB).Scan(&locked)
	if err != nil || !locked {
		t.Fatalf("lock=%t err=%v", locked, err)
	}
	err = ProvisionTenant(ctx, cfg, tenant)
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatal(err)
	}
	if _, err = conn.Exec(ctx, "SELECT pg_catalog.pg_advisory_unlock($1, $2)", keyA, keyB); err != nil {
		t.Fatal(err)
	}
	if err = ProvisionTenant(ctx, cfg, tenant); err != nil {
		t.Fatal(err)
	}
}

func TestTenantJobCredentialIsSchemaLocal(t *testing.T) {
	ctx := context.Background()
	cfg := gooseConfig("goose_tenant_ddl")
	if err := foundation.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := UpPlatform(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	first := [16]byte{2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5}
	second := [16]byte{5, 5, 5, 5, 4, 4, 4, 4, 3, 3, 3, 3, 2, 2, 2, 2}
	if err := ProvisionTenant(ctx, cfg, first); err != nil {
		t.Fatal(err)
	}
	if err := ProvisionTenant(ctx, cfg, second); err != nil {
		t.Fatal(err)
	}
	leftSchema, err := foundation.TenantSchemaName(first)
	if err != nil {
		t.Fatal(err)
	}
	rightSchema, err := foundation.TenantSchemaName(second)
	if err != nil {
		t.Fatal(err)
	}
	role, err := foundation.TenantRoleName(first)
	if err != nil {
		t.Fatal(err)
	}
	const password = "tenant-job-isolation-secret"
	if err = setRolePassword(t, ctx, cfg, role, password); err != nil {
		t.Fatal(err)
	}
	dsn, err := userDSN(cfg.AdminDSN, role, password, cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	job, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer job.Close(ctx)
	own := "CREATE TABLE " + pgx.Identifier{leftSchema}.Sanitize() + ".job_probe(id integer)"
	if _, err = job.Exec(ctx, own); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE platform.job_probe(id integer)",
		"CREATE TABLE " + pgx.Identifier{rightSchema}.Sanitize() + ".job_probe(id integer)",
		"SELECT id FROM platform.tenants",
		"CREATE SCHEMA job_extra",
	} {
		if _, err = job.Exec(ctx, statement); err == nil {
			t.Fatalf("tenant job credential ran %s", statement)
		}
	}
}
