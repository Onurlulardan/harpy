//go:build integration

package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

const integrationCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func OpenIntegrationAPI(
	ctx context.Context,
	fixture configfixture.APISetter,
	dsn string,
	extensionManifestSHA256 string,
) (*Pool, error) {
	if fixture == nil || dsn == "" {
		return nil, fmt.Errorf("PostgreSQL integration API fixture is invalid")
	}
	fixture.Helper()
	fixture.Setenv("HARPY_ENVIRONMENT", "test")
	fixture.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	fixture.Setenv("HARPY_POSTGRES_DSN", dsn)
	fixture.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", integrationCursorKeyring)
	fixture.Setenv("HARPY_POSTGRES_MAX_CONNS", "4")
	fixture.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	fixture.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(fixture)
	loaded, err := config.LoadAPI()
	if err != nil {
		return nil, fmt.Errorf("load PostgreSQL integration API config: %w", err)
	}
	contract, err := schemacontract.IntegrationContract(extensionManifestSHA256)
	if err != nil {
		return nil, err
	}
	pool, err := Open(
		ctx,
		loaded.LongRunning().Runtime().Postgres(),
		ProfileAPI,
		contract,
	)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL integration API pool: %w", err)
	}
	return pool, nil
}
