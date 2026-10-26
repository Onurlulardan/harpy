//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/jackc/pgx/v5/stdlib"
)

type pipelineWorkerRuntime struct {
	coordinator *platformtx.Coordinator
	rabbit      config.RabbitMQ
}

func preparePipelineRuntime(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	rabbitURL string,
) pipelineWorkerRuntime {
	t.Helper()
	preparePipelineRoles(t, ctx, fixture.owner)
	preparePipelineTenant(t, ctx, fixture.owner)
	preparePipelineFoundation(t, ctx, fixture.owner)
	assertPipelineRuntimeGrants(t, ctx, fixture.runtime)
	runtimeDSN := pipelineRuntimeDSN(t, ctx, fixture.owner)
	loaded := loadPipelineWorkerConfig(t, runtimeDSN, rabbitURL)
	pool, err := platformpostgres.Open(
		ctx, loaded.LongRunning().Runtime().Postgres(), platformpostgres.ProfileWorker,
		pipelineSchemaContract(t),
	)
	if err != nil {
		t.Fatalf("open production worker pool: %v", err)
	}
	t.Cleanup(pool.Close)
	assertPipelineScopePrerequisites(t, ctx, pool)
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		t.Fatal(err)
	}
	return pipelineWorkerRuntime{coordinator: coordinator, rabbit: loaded.RabbitMQ()}
}

func loadPipelineWorkerConfig(
	t *testing.T,
	runtimeDSN string,
	rabbitURL string,
) config.Worker {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", runtimeDSN)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "4")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	t.Setenv("HARPY_RABBITMQ_URL", rabbitURL)
	configfixture.SetWorker(t)
	loaded, err := config.LoadWorker()
	if err != nil {
		t.Fatalf("load pipeline worker config: %v", err)
	}
	return loaded
}

func pipelineRuntimeDSN(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) string {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var dsn string
	err = connection.Raw(func(raw any) error {
		stdlibConnection, ok := raw.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("unexpected SQL connection %T", raw)
		}
		parsed, err := url.Parse(stdlibConnection.Conn().Config().ConnString())
		if err != nil {
			return err
		}
		parsed.User = url.UserPassword("harpy_v2_runtime", "runtime_integration_only")
		dsn = parsed.String()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return dsn
}
