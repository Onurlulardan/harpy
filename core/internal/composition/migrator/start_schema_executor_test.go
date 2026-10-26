package migrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestStartSchemaExecutorBoundsAndCancelsDatabaseStartup(t *testing.T) {
	setSchemaExecutorTestEnvironment(t)
	t.Setenv("HARPY_READINESS_TIMEOUT", "250ms")
	settings, err := config.LoadSchemaExecutor()
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
	code := startSchemaExecutorWithOpen(
		lifecycle.LocalProcess(parent), settings, buildinfo.Info{},
		schemacontract.EmbeddedContract{}, nil, nil, opener,
	)
	if code != processcode.DependencyPreflight || !deadlineSeen {
		t.Fatalf("code=%d deadline_seen=%t", code, deadlineSeen)
	}
}
