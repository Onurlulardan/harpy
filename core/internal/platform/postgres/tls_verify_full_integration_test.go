//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/jackc/pgx/v5"
)

func TestProductionPostgresVerifyFullAgainstExactPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	instance := startPostgresTLSInstance(t, ctx)
	trustedDSN := instance.dsn("localhost", "verify-full", instance.authority.caPath)
	settings, err := config.IntegrationLoadAPI(postgresTLSProductionEnvironment(t, trustedDSN))
	if err != nil {
		t.Fatalf("load production verify-full configuration: %v", err)
	}
	connection, err := pgx.Connect(
		ctx, settings.LongRunning().Runtime().Postgres().DSN().Value(),
	)
	if err != nil {
		t.Fatalf("connect with trusted CA and exact hostname: %v", err)
	}
	defer connection.Close(context.Background())
	var version int
	var encrypted bool
	if err := connection.QueryRow(ctx, `SELECT
current_setting('server_version_num')::integer,
ssl FROM pg_catalog.pg_stat_ssl WHERE pid=pg_catalog.pg_backend_pid()`).Scan(
		&version, &encrypted,
	); err != nil {
		t.Fatal(err)
	}
	if version != runtimev1.PostgresServerVersionNumber || !encrypted {
		t.Fatalf("unexpected TLS PostgreSQL runtime: version=%d ssl=%t", version, encrypted)
	}
	assertPostgresTLSConnectionRejected(
		t, ctx, instance.dsn("localhost", "verify-full", instance.wrongCA),
	)
	assertPostgresTLSConnectionRejected(
		t, ctx, instance.dsn("127.0.0.1", "verify-full", instance.authority.caPath),
	)
	for _, mode := range []string{"disable", "require", "verify-ca"} {
		t.Run("reject production "+mode, func(t *testing.T) {
			if _, err := config.IntegrationLoadAPI(postgresTLSProductionEnvironment(t,
				instance.dsn("localhost", mode, ""),
			)); err == nil {
				t.Fatalf("production accepted weak PostgreSQL TLS mode %s", mode)
			}
		})
	}
}

func assertPostgresTLSConnectionRejected(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	settings, err := config.IntegrationLoadAPI(postgresTLSProductionEnvironment(t, dsn))
	if err != nil {
		t.Fatalf("negative TLS configuration should pass structural validation: %v", err)
	}
	connection, err := pgx.Connect(
		ctx, settings.LongRunning().Runtime().Postgres().DSN().Value(),
	)
	if connection != nil {
		_ = connection.Close(context.Background())
	}
	if err == nil {
		t.Fatal("PostgreSQL verify-full accepted untrusted certificate identity")
	}
}

func postgresTLSProductionEnvironment(t *testing.T, dsn string) []string {
	t.Helper()
	entries := []string{
		"HARPY_ENVIRONMENT=production",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=" + dsn,
		"HARPY_OTLP_ENDPOINT=https://collector.invalid:4318",
		"HARPY_REST_QUERY_CURSOR_KEYRING=" + runtimeTestCursorKeyring,
	}
	return append(entries, configfixture.APIEnvironment(t, time.Now().UTC())...)
}
