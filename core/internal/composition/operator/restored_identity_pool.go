package operator

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func openIdentityReconcilePool(
	ctx context.Context,
	runtime config.Runtime,
) (*platformpostgres.Pool, error) {
	connectContext, cancel := context.WithTimeout(ctx,
		runtime.Postgres().ConnectTimeout()+runtime.ReadinessTimeout()+5*time.Second)
	defer cancel()
	return platformpostgres.Open(connectContext, runtime.Postgres(),
		platformpostgres.ProfileOperatorIdentityReconcile)
}
