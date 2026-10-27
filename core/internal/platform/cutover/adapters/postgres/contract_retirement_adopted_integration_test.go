//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func verifyContractRetirementAdoptedMigrations(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	if _, err := database.admin.Exec(ctx,
		"CREATE DATABASE harpy_cutover_adopted OWNER harpy_owner"); err != nil {
		t.Fatal(err)
	}
	adminConfig, err := pgx.ParseConfig(database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig.Database = "harpy_cutover_adopted"
	admin, err := pgx.ConnectConfig(ctx, adminConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	migratorConfig := adminConfig.Copy()
	migratorConfig.User = "harpy_migrator"
	migratorConfig.Password = "integration-migrator-secret"
	migrator, err := pgx.ConnectConfig(ctx, migratorConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer migrator.Close(ctx)
	if _, err := migrator.Exec(ctx, "SET ROLE harpy_owner; CREATE SCHEMA platform"); err != nil {
		t.Fatal(err)
	}
	prepareReadinessSurfaces(t, ctx, migrator, "fresh")
	applyContractRetirementMigrationVariant(t, ctx, migrator, "fresh")
	assertCutoverFunctionAuthority(t, ctx, admin)
}

func applyContractRetirementMigrationVariant(
	t *testing.T, ctx context.Context, connection *pgx.Conn, variant string,
) {
	t.Helper()
	if variant != "fresh" {
		t.Fatal("fresh contract retirement migrations did not load")
	}
	directory := filepath.Join("..", "..", "migrations", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	suffix := "." + variant + ".sql"
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		document, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := connection.Exec(ctx, string(document)); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) {
				t.Fatalf("apply %s: SQLSTATE=%s position=%d: %v",
					name, postgresError.Code, postgresError.Position, err)
			}
			t.Fatalf("apply %s: %v", name, err)
		}
	}
}
