//go:build integration

package initialize

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const initializeImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

var adminDSN string

func TestMain(m *testing.M) {
	os.Exit(runInitializeIntegration(m))
}

func runInitializeIntegration(m *testing.M) int {
	if dsn := os.Getenv("HARPY_FOUNDATION_TEST_ADMIN_DSN"); dsn != "" {
		adminDSN = dsn
		return m.Run()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := tcpostgres.Run(
		ctx, initializeImage,
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
		fmt.Fprintf(os.Stderr, "start PostgreSQL 18 initialize database: %v\n", err)
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(container) }()
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize database dsn: %v\n", err)
		return 1
	}
	adminDSN = dsn
	return m.Run()
}

func TestEmptyDatabaseInitializeRunsFourStagesOnce(t *testing.T) {
	cfg := integrationConfig(t)
	calls := []string{}
	result, err := Run(context.Background(), Request{
		Config: cfg,
		Goose: func(context.Context) error {
			calls = append(calls, "goose")
			return nil
		},
		Application: func(context.Context) error {
			calls = append(calls, "application")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stages) != 4 || strings.Join(calls, ",") != "goose,application" {
		t.Fatalf("stages=%#v calls=%v", result.Stages, calls)
	}
	report := FormatReport(cfg.Database, result)
	if strings.Count(report, "status=ok") != 4 || strings.Contains(report, cfg.GoosePassword) {
		t.Fatal(report)
	}
	before := relationCount(t, cfg.Database)
	if _, err := Run(context.Background(), Request{
		Config: cfg,
		Goose: func(context.Context) error {
			calls = append(calls, "goose")
			return nil
		},
		Application: func(context.Context) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	if relationCount(t, cfg.Database) != before {
		t.Fatal("second initialize created relations")
	}
}

func TestDirtyTargetStopsBeforeGoose(t *testing.T) {
	cfg := integrationConfig(t)
	admin := openAdmin(t)
	if _, err := admin.Exec(context.Background(), "CREATE DATABASE "+pgx.Identifier{cfg.Database}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	target := openDatabase(t, adminDSN, cfg.Database)
	if _, err := target.Exec(context.Background(), "CREATE TABLE public.legacy(id integer)"); err != nil {
		t.Fatal(err)
	}
	if _, err := target.Exec(context.Background(), "INSERT INTO public.legacy(id) VALUES (4)"); err != nil {
		t.Fatal(err)
	}
	goose := false
	_, err := Run(context.Background(), Request{
		Config:      cfg,
		Goose:       func(context.Context) error { goose = true; return nil },
		Application: func(context.Context) error { goose = true; return nil },
	})
	if err == nil || goose || !strings.Contains(err.Error(), "public.legacy") {
		t.Fatalf("goose=%t err=%v", goose, err)
	}
	var id int
	if err := target.QueryRow(context.Background(), "SELECT id FROM public.legacy").Scan(&id); err != nil || id != 4 {
		t.Fatal(err)
	}
}

func TestGooseStageRetriesWithoutWritingHistory(t *testing.T) {
	cfg := integrationConfig(t)
	attempts := 0
	started := false
	_, err := Run(context.Background(), Request{
		Config: cfg,
		Goose: func(context.Context) error {
			attempts++
			return fmt.Errorf("SQLSTATE 42P07 %s", cfg.GoosePassword)
		},
		Application: func(context.Context) error { started = true; return nil },
	})
	if err == nil || started || !strings.Contains(err.Error(), "SQLSTATE 42P07") ||
		strings.Contains(err.Error(), cfg.GoosePassword) {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), Request{
		Config: cfg,
		Goose: func(context.Context) error {
			attempts++
			return nil
		},
		Application: func(context.Context) error { started = true; return nil },
	})
	if err != nil || !started || attempts != 2 {
		t.Fatalf("attempts=%d started=%t err=%v", attempts, started, err)
	}
	if relationCount(t, cfg.Database) != 0 {
		t.Fatal("goose retry created application relations or history")
	}
}

func integrationConfig(t *testing.T) foundation.Config {
	t.Helper()
	name := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	if len(name) > 40 {
		name = name[:40]
	}
	return foundation.Config{
		AdminDSN: adminDSN, Database: "hit_" + name,
		GoosePassword: "goose-initialize-test-secret", APIPassword: "api-initialize-test-secret01",
		WorkerPassword: "worker-initialize-test-secret",
	}
}
