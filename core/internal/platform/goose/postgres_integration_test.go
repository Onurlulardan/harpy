//go:build integration

package goose

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const gooseImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

var gooseAdminDSN string

func TestMain(m *testing.M) {
	os.Exit(runGooseIntegration(m))
}

func runGooseIntegration(m *testing.M) int {
	if dsn := os.Getenv("HARPY_FOUNDATION_TEST_ADMIN_DSN"); dsn != "" {
		gooseAdminDSN = dsn
		release, err := foundation.LockSharedTestCluster(dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lock shared foundation database: %v\n", err)
			return 1
		}
		defer release()
		return m.Run()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(
		ctx, gooseImage,
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
		fmt.Fprintf(os.Stderr, "start PostgreSQL 18 goose database: %v\n", err)
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(container) }()
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "goose database dsn: %v\n", err)
		return 1
	}
	gooseAdminDSN = dsn
	return m.Run()
}

func gooseConfig(database string) foundation.Config {
	return foundation.Config{
		AdminDSN: gooseAdminDSN, Database: database,
		GoosePassword:  "goose-foundation-test-secret",
		APIPassword:    "api-foundation-test-secret01",
		WorkerPassword: "worker-foundation-test-secret",
	}
}
