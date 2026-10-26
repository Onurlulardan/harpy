//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type iamPlatformGuardDatabase struct {
	dsn      string
	admin    *sql.DB
	migrator *sql.DB
	runtime  *sql.DB
	adopter  *sql.DB
}

func startIAMPlatformGuardPG18(
	t *testing.T, ctx context.Context, origin string,
) iamPlatformGuardDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, iamFoundationPGImage,
		tcpostgres.WithDatabase("harpy_iam_guard_"+origin),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("iam-guard-admin-only"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256",
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=scram-sha-256 --auth-host=scram-sha-256",
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
		testcontainers.WithCmd("postgres"),
	)
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start IAM guard PostgreSQL 18.4: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate IAM guard database: %v", err)
		}
	})
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	admin := openIAMPlatformGuardRole(t, dsn, "postgres", "iam-guard-admin-only", "")
	prepareIAMPlatformGuardRoles(t, ctx, admin)
	database := iamPlatformGuardDatabase{
		dsn:      dsn,
		admin:    admin,
		migrator: openIAMPlatformGuardRole(t, dsn, "harpy_migrator", "guard-migrator", "harpy_owner"),
		runtime:  openIAMPlatformGuardRole(t, dsn, "harpy_v2_runtime", "guard-runtime", "harpy_runtime"),
		adopter:  openIAMPlatformGuardRole(t, dsn, "harpy_adopter", "guard-adopter", ""),
	}
	for _, connection := range []*sql.DB{
		database.admin, database.migrator, database.runtime, database.adopter,
	} {
		t.Cleanup(func() { _ = connection.Close() })
	}
	var version int
	if err := admin.QueryRowContext(ctx,
		"SELECT current_setting('server_version_num')::integer").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 180004 {
		t.Fatalf("PostgreSQL version = %d, want 180004", version)
	}
	return database
}

func openIAMPlatformGuardRole(
	t *testing.T, dsn, user, password, role string,
) *sql.DB {
	t.Helper()
	configuration, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.User, configuration.Password = user, password
	if role != "" {
		configuration.RuntimeParams["role"] = role
	}
	return stdlib.OpenDB(*configuration)
}

func prepareIAMPlatformGuardRoles(
	t *testing.T, ctx context.Context, admin *sql.DB,
) {
	t.Helper()
	_, err := admin.ExecContext(ctx, `
CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD 'guard-migrator';
GRANT harpy_owner TO harpy_migrator WITH INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_restore NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0 PASSWORD NULL;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD 'guard-runtime';
GRANT harpy_runtime TO harpy_v2_runtime
WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_adopter LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD 'guard-adopter';
DO $do$
BEGIN
    EXECUTE pg_catalog.format(
        'GRANT CREATE ON DATABASE %I TO harpy_owner',
        pg_catalog.current_database()
    );
END;
$do$;`)
	if err != nil {
		t.Fatalf("prepare IAM guard roles: %v", err)
	}
}
