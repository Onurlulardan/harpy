//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const localReleaseIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func TestPrepareProvisionsCanonicalPostgresAndReusesOwnedOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(
		ctx, localReleaseIntegrationImage,
		tcpostgres.WithDatabase("postgres"), tcpostgres.WithUsername("local_admin"),
		tcpostgres.WithPassword("container-admin-only"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256",
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 --data-checksums " +
				"--auth-local=trust --auth-host=scram-sha-256",
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
		testcontainers.WithCmd("postgres"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	adminPath := writeIntegrationSecret(t, directory, "admin", dsn)
	runtimePath := writeIntegrationSecret(t, directory, "runtime", "runtime-integration-only")
	migratorPath := writeIntegrationSecret(t, directory, "migrator", "migrator-integration-only")
	backupPath := writeIntegrationSecret(t, directory, "backup", "backup-integration-only")
	restorePath := writeIntegrationSecret(t, directory, "restore", "restore-integration-only")
	observerPath := writeIntegrationSecret(t, directory, "observer", "observer-integration-only")
	migration, err := loadMigrationArtifact()
	if err != nil {
		t.Fatal(err)
	}
	migrationPath := filepath.Join(directory, "migration-manifest.json")
	if err := os.WriteFile(migrationPath, migration.document, 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "release")
	arguments := []string{
		"-admin-dsn-file", adminPath, "-runtime-password-file", runtimePath,
		"-migrator-password-file", migratorPath, "-backup-password-file", backupPath,
		"-restore-password-file", restorePath, "-migration-manifest", migrationPath,
		"-reactivation-observer-password-file", observerPath,
		"-output", output,
	}
	options, err := parsePrepareOptions(arguments)
	if err != nil {
		t.Fatal(err)
	}
	if err := executePrepare(ctx, options); err == nil {
		t.Fatal("prepare still emits a migration manifest")
	}
}

func writeIntegrationSecret(t *testing.T, directory, name, value string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func createIntegrationReadinessTables(
	t *testing.T, ctx context.Context, dsn, output string,
) {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(context.Background())
	tx, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "CREATE SCHEMA platform AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE platform.environment_identity (
singleton boolean PRIMARY KEY, environment_id uuid NOT NULL, database_origin text NOT NULL,
database_system_identifier text NOT NULL, database_oid bigint NOT NULL,
release_id uuid NOT NULL, bootstrap_authority_sha256 text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	identity := readIntegrationEvidenceIdentity(t, output)
	if _, err := tx.Exec(ctx, `INSERT INTO platform.environment_identity VALUES
(TRUE,$1,'fresh',$2,$3,$4,$5)`, identity.environmentID, identity.systemID,
		identity.databaseOID, identity.releaseID, identity.authorityDigest); err != nil {
		t.Fatal(err)
	}
	for _, table := range runtimeReadinessTables[1:] {
		statement := "CREATE TABLE " + pgx.Identifier{"platform", table}.Sanitize() + " (id bigint)"
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
