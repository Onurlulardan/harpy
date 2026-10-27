package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestStartWorkerBoundsAndCancelsDatabaseStartup(t *testing.T) {
	setWorkerStartupEnvironment(t)
	settings, err := config.LoadWorker()
	if err != nil {
		t.Fatal(err)
	}
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	deadlineSeen := false
	opener := func(
		ctx context.Context,
		_ config.Postgres,
		_ postgres.Profile,
		_ schemacontract.EmbeddedContract,
	) (*postgres.Pool, error) {
		deadline, present := ctx.Deadline()
		deadlineSeen = present && time.Until(deadline) <= time.Second
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("startup context error = %v", ctx.Err())
		}
		return nil, ctx.Err()
	}
	code := startWorkerWithOpen(
		lifecycle.LocalProcess(parent), settings, buildinfo.Info{},
		schemacontract.EmbeddedContract{}, nil, nil, opener,
	)
	if code != processcode.DependencyPreflight || !deadlineSeen {
		t.Fatalf("code=%d deadline_seen=%t", code, deadlineSeen)
	}
}

func setWorkerStartupEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_runtime@127.0.0.1:1/harpy?sslmode=disable",
	)
	t.Setenv("HARPY_READINESS_TIMEOUT", "250ms")
	t.Setenv("HARPY_RABBITMQ_URL", "amqp://worker:secret@127.0.0.1:1/%2f")
	configfixture.SetWorker(t)
}
