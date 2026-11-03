//go:build integration

package foundation

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const foundationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const (
	goosePassword  = "goose-foundation-test-secret"
	apiPassword    = "api-foundation-test-secret01"
	workerPassword = "worker-foundation-test-secret"
)

var adminDSN string

func TestMain(m *testing.M) {
	os.Exit(runIntegration(m))
}

func runIntegration(m *testing.M) int {
	if dsn := os.Getenv("HARPY_FOUNDATION_TEST_ADMIN_DSN"); dsn != "" {
		adminDSN = dsn
		return m.Run()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(
		ctx, foundationImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("foundation-admin-test-secret"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(2*time.Minute)),
		testcontainers.WithCmd("postgres"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start PostgreSQL 18 foundation database: %v\n", err)
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(container) }()
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "foundation database dsn: %v\n", err)
		return 1
	}
	adminDSN = dsn
	return m.Run()
}

func testConfig(database string) Config {
	return Config{
		AdminDSN: adminDSN, Database: database,
		GoosePassword: goosePassword, APIPassword: apiPassword, WorkerPassword: workerPassword,
	}
}

func databaseName(t *testing.T) string {
	t.Helper()
	var name strings.Builder
	name.WriteString("hit_")
	for _, character := range strings.ToLower(t.Name()) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			name.WriteRune(character)
		} else {
			name.WriteByte('_')
		}
	}
	value := strings.TrimRight(name.String(), "_")
	if len(value) > 48 {
		value = strings.TrimRight(value[:48], "_")
	}
	return value
}

func adminConn(t *testing.T) (*pgx.Conn, context.Context) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	conn, err := connect(ctx, adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	if err := resetCluster(ctx, conn); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		cleanup, err := connect(cleanupCtx, adminDSN)
		if err != nil {
			return
		}
		defer cleanup.Close(cleanupCtx)
		_ = resetCluster(cleanupCtx, cleanup)
	})
	return conn, ctx
}

func resetCluster(ctx context.Context, conn *pgx.Conn) error {
	rows, err := conn.Query(ctx, `SELECT datname FROM pg_catalog.pg_database
WHERE datname LIKE 'hit\_%' ESCAPE '\'`)
	if err != nil {
		return err
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	for _, name := range names {
		_, _ = conn.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		if _, err := conn.Exec(ctx, "DROP DATABASE "+ident(name)); err != nil {
			return err
		}
	}
	_, _ = conn.Exec(ctx, "REVOKE "+ident(RoleOwner)+" FROM "+ident(RoleGoose))
	roleRows, err := conn.Query(ctx, `SELECT rolname FROM pg_catalog.pg_roles
WHERE rolname LIKE 'harpy\_%' ESCAPE '\' ORDER BY rolname DESC`)
	if err != nil {
		return err
	}
	roles, err := pgx.CollectRows(roleRows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	for _, role := range roles {
		if _, err := conn.Exec(ctx, "DROP ROLE IF EXISTS "+ident(role)); err != nil {
			return err
		}
	}
	return nil
}

func adminIdentity(t *testing.T) (string, string) {
	t.Helper()
	parsed, err := url.Parse(adminDSN)
	if err != nil || parsed.User == nil {
		t.Fatal(err)
	}
	password, ok := parsed.User.Password()
	if !ok || password == "" {
		t.Fatal("admin dsn has no password")
	}
	return parsed.User.Username(), password
}

func connectRole(t *testing.T, ctx context.Context, user, password, database string) *pgx.Conn {
	t.Helper()
	dsn, err := userDSN(adminDSN, user, password, database)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}
