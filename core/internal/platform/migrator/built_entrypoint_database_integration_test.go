//go:build integration

package migrator

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const builtEntrypointDatabaseName = "harpy_entrypoint_gate"

const builtEntrypointRolePassword = "entrypoint-role-only"

type builtEntrypointDatabase struct {
	admin       *pgx.Conn
	runtimeDSN  string
	migratorDSN string
}

func startBuiltEntrypointDatabase(
	t *testing.T,
	ctx context.Context,
) builtEntrypointDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, runnerIntegrationImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("entrypoint_test_admin"),
		tcpostgres.WithPassword("entrypoint-container-only"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256",
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=trust --auth-host=scram-sha-256",
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
		t.Fatalf("start built entrypoint PostgreSQL 18.4: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate built entrypoint PostgreSQL: %v", err)
		}
	})
	adminDSN, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	clusterAdmin, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	prepareBuiltEntrypointCluster(t, ctx, clusterAdmin)
	assertProductionRoleFixture(t, ctx, clusterAdmin)
	if _, err := clusterAdmin.Exec(ctx, `ALTER ROLE harpy_v2_runtime
PASSWORD 'entrypoint-role-only'; ALTER ROLE harpy_migrator
PASSWORD 'entrypoint-role-only'`); err != nil {
		t.Fatalf("lease built entrypoint test roles: %v", err)
	}
	_ = clusterAdmin.Close(context.Background())
	targetURL, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	targetURL.Path = "/" + builtEntrypointDatabaseName
	admin, err := pgx.Connect(ctx, targetURL.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	prepareBuiltEntrypointExtensions(t, ctx, admin)
	runtimeDSN := builtEntrypointRoleDSN(t, targetURL, "harpy_v2_runtime")
	migratorDSN := builtEntrypointRoleDSN(t, targetURL, "harpy_migrator")
	return builtEntrypointDatabase{
		admin: admin, runtimeDSN: runtimeDSN, migratorDSN: migratorDSN,
	}
}

func builtEntrypointRoleDSN(t *testing.T, endpoint *url.URL, role string) string {
	t.Helper()
	copy := *endpoint
	copy.User = url.UserPassword(role, builtEntrypointRolePassword)
	return copy.String()
}
