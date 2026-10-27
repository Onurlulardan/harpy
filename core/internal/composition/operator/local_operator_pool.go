package operator

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func openLocalOperatorPool(
	ctx context.Context,
	runtime config.Runtime,
) (*platformpostgres.Pool, error) {
	return openOperatorPool(ctx, runtime, platformpostgres.ProfileOperatorInitialize)
}

func openOperatorPool(
	ctx context.Context,
	runtime config.Runtime,
	profile platformpostgres.Profile,
) (*platformpostgres.Pool, error) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return nil, err
	}
	embedded, err := schemacontract.LoadEmbedded(registry)
	if err != nil {
		return nil, err
	}
	connectContext, cancel := context.WithTimeout(ctx,
		runtime.Postgres().ConnectTimeout()+runtime.ReadinessTimeout()+5*time.Second)
	defer cancel()
	return platformpostgres.Open(connectContext, runtime.Postgres(),
		profile, embedded)
}
