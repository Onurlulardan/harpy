package api

import (
	"context"
	"fmt"

	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildCustomerProvisioningAuthority(
	ctx context.Context,
	settings config.API,
	pool *platformpostgres.Pool,
	policy *tenancymodule.DeploymentPolicyRuntime,
	clock customerProvisioningClock,
) (tenancymodule.CustomerProvisioningAuthorityPlanner, *platformauthority.GoogleClient, error) {
	mode, err := settings.DeploymentMode()
	if err != nil {
		return nil, nil, err
	}
	if mode != config.DeploymentSaaS {
		return customerProvisioningAuthority{}, nil, nil
	}
	if err := policy.Verify(ctx); err != nil {
		return nil, nil, fmt.Errorf("verify SaaS provisioning policy: %w", err)
	}
	snapshot, err := policy.Snapshot()
	if err != nil {
		return nil, nil, err
	}
	binding, err := platformauthority.ResolveSaaSBinding(
		ctx, pool, clock, snapshot.EnvironmentID(),
	)
	if err != nil {
		return nil, nil, err
	}
	authorityConfig := settings.Authority()
	if authorityConfig.ProposerKMSKeyID().Value() != binding.SignerKID {
		return nil, nil, fmt.Errorf("configured SaaS authority signer differs from persisted binding")
	}
	identityPath := ""
	if identity, configured := authorityConfig.WorkloadIdentity(); configured {
		identityPath = identity.Path().Value()
	}
	client, err := platformauthority.OpenGoogleClient(ctx, identityPath)
	if err != nil {
		return nil, nil, err
	}
	planner, err := platformauthority.NewSaaSPlanner(ctx,
		platformauthority.SaaSPlannerDependencies{Pool: pool, Clock: clock,
			KMSKeyID:           binding.SignerKID,
			KMSPublicKeySHA256: binding.SignerPublicKeySHA256, KMSClient: client})
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	return customerProvisioningAuthority{planner: planner, binding: binding}, client, nil
}
