package platformauthority

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/gcpkms"
	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
)

func NewSaaSPlanner(
	ctx context.Context,
	dependencies SaaSPlannerDependencies,
) (api.InitialAuthorityPlanner, error) {
	if dependencies.Clock == nil {
		return nil, fmt.Errorf("SaaS platform authority planner clock is missing")
	}
	signer, err := gcpkms.NewSigner(ctx, dependencies.KMSKeyID,
		dependencies.KMSPublicKeySHA256, dependencies.KMSClient)
	if err != nil {
		return nil, err
	}
	store, err := postgresadapter.NewStore(dependencies.Pool, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	return application.NewInitialPlanner(store, signer, dependencies.Clock)
}
