//go:build integration

package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func runtimePoolConfig(t *testing.T, dsn string) config.Postgres {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", dsn)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", runtimeTestCursorKeyring)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "2")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatalf("load runtime pool config: %v", err)
	}
	return loaded.LongRunning().Runtime().Postgres()
}

const runtimeExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const runtimeTestCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`
