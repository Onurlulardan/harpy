package operator

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
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
	embedded, err := schemacontract.LoadBoundStartup()
	if err != nil {
		return nil, err
	}
	connectContext, cancel := context.WithTimeout(ctx,
		runtime.Postgres().ConnectTimeout()+runtime.ReadinessTimeout()+5*time.Second)
	defer cancel()
	return platformpostgres.Open(connectContext, runtime.Postgres(),
		profile, embedded)
}
