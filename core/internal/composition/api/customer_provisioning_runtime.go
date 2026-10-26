package api

import (
	"context"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type customerProvisioningRuntime struct {
	http      *tenancymodule.CustomerProvisioningHTTP
	kmsClient *platformauthority.GoogleClient
}

func (runtime *customerProvisioningRuntime) Routes() []httpedge.Route {
	if runtime == nil || runtime.http == nil {
		return nil
	}
	return runtime.http.Routes()
}

func (runtime *customerProvisioningRuntime) Close() {
	if runtime != nil && runtime.kmsClient != nil {
		_ = runtime.kmsClient.Close()
	}
}

func buildCustomerProvisioningRuntime(
	ctx context.Context,
	settings config.API,
	pool *platformpostgres.Pool,
	policy *tenancymodule.DeploymentPolicyRuntime,
	access *iammodule.CustomerProvisioningAccessRuntime,
	delivery *iammodule.DeliverySealer,
	report httpedge.ErrorSink,
) (*customerProvisioningRuntime, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil || policy == nil ||
		access == nil || delivery == nil || report == nil {
		return nil, fmt.Errorf("customer provisioning dependencies are incomplete")
	}
	store, err := tenancymodule.NewCustomerProvisioningStore(pool)
	if err != nil {
		return nil, fmt.Errorf("build customer provisioning store: %w", err)
	}
	reader, err := billingmodule.NewSaaSFreeBasePlatformReader(pool)
	if err != nil {
		return nil, fmt.Errorf("build SaaS base offering reader: %w", err)
	}
	iamSettings, err := settings.IAM()
	if err != nil {
		return nil, err
	}
	setup, err := iammodule.NewInitialAdminSetupPreparer(
		pool, iamSettings, delivery,
	)
	if err != nil {
		return nil, fmt.Errorf("build initial admin setup preparer: %w", err)
	}
	reissuer, err := iammodule.NewInitialAdminSetupReissuer(pool, setup)
	if err != nil {
		return nil, fmt.Errorf("build initial admin setup reissuer: %w", err)
	}
	clock := customerProvisioningClock{}
	authority, client, err := buildCustomerProvisioningAuthority(
		ctx, settings, pool, policy, clock,
	)
	if err != nil {
		return nil, err
	}
	coordinator, err := tenancymodule.NewCustomerProvisioningApplication(tenancymodule.CustomerProvisioningDependencies{
		Policy: policy, Access: customerProvisioningAccess{runtime: access},
		Offering: customerProvisioningOffering{reader: reader}, Authority: authority,
		Setup: customerProvisioningSetup{preparer: setup, reissuer: reissuer},
		Store: store, Clock: clock,
	})
	if err != nil {
		if client != nil {
			_ = client.Close()
		}
		return nil, err
	}
	handler, err := tenancymodule.NewCustomerProvisioningHTTP(coordinator, report)
	if err != nil {
		if client != nil {
			_ = client.Close()
		}
		return nil, err
	}
	return &customerProvisioningRuntime{http: handler, kmsClient: client}, nil
}
