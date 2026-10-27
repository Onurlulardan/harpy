package migrator

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestPrepareDeployConnectionPinsMigratorSessionWithoutConnecting(t *testing.T) {
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_migrator@127.0.0.1:1/harpy?sslmode=disable",
	)
	settings, err := config.LoadSchemaExecutor()
	if err != nil {
		t.Fatalf("load database settings: %v", err)
	}
	connection, err := prepareDeployConnection(settings.LongRunning().Runtime().Postgres())
	if err != nil {
		t.Fatalf("prepare deploy connection: %v", err)
	}
	if connection.Database != "harpy" || connection.ConnectTimeout != 5*time.Second {
		t.Fatalf("unexpected deploy target: database=%q timeout=%s", connection.Database, connection.ConnectTimeout)
	}
	for key, expected := range map[string]string{
		"role": "harpy_owner", "search_path": "pg_catalog",
		"application_name": "harpy-v2:core-migrator-deploy",
	} {
		if connection.RuntimeParams[key] != expected {
			t.Fatalf("runtime parameter %s=%q, want %q", key, connection.RuntimeParams[key], expected)
		}
	}
}
