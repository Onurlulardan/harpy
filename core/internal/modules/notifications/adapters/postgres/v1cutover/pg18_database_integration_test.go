//go:build integration

package v1cutover

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const cutoverPG18Image = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

func startCutoverPG18(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()
	container, err := tcpostgres.Run(ctx, cutoverPG18Image,
		tcpostgres.WithDatabase("harpy_notification_cutover"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("notification_cutover_only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp")).WithStartupTimeout(90*time.Second)),
		testcontainers.WithEnv(map[string]string{"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 --data-checksums " +
			"--auth-local=trust --auth-host=scram-sha-256"}),
		testcontainers.WithCmd("postgres"))
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact PostgreSQL 18.4: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close(ctx) })
	var raw string
	if err = connection.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		t.Fatal(err)
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version != 180004 {
		t.Fatalf("PostgreSQL version=%d err=%v", version, err)
	}
	return connection
}
