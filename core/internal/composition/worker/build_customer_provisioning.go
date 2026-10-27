package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type customerWorkerClock struct{}

func (customerWorkerClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func buildCustomerProvisioningWorkload(
	ctx context.Context,
	pool *platformpostgres.Pool,
	settings config.Worker,
	iamSettings config.IAMWorker,
) (managedAsyncWorkload, bool, error) {
	mode, err := settings.DeploymentMode()
	if err != nil {
		return nil, false, err
	}
	if mode != config.DeploymentSaaS {
		return nil, false, nil
	}
	if ctx == nil || ctx.Err() != nil || pool == nil {
		return nil, false, fmt.Errorf("customer provisioning worker dependencies are invalid")
	}
	authority := settings.Authority()
	route, err := routeAuthorityWorkload(authority.Mode())
	if err != nil || route != authorityRouteSaaSRegistry {
		return nil, false, fmt.Errorf("customer provisioning requires SaaS authority")
	}
	clock := customerWorkerClock{}
	ackTrust, err := readCustomerAuthorityArtifact(authority.AckTrustBundle().Path().Value())
	if err != nil {
		return nil, false, err
	}
	defer clear(ackTrust)
	releaseRoot, err := selectCustomerReleaseRoot(
		authority.ReleaseTrustRoot().Path().Value(), ackTrust, clock.Now(),
	)
	if err != nil {
		return nil, false, err
	}
	acknowledger, err := platformauthority.NewSaaSAcknowledger(
		platformauthority.SaaSAcknowledgerDependencies{
			Pool: pool, Clock: clock, HTTPClient: &http.Client{Timeout: 15 * time.Second},
			RegistryURL: authority.RegistryURL().Value(), AckTrustBundle: ackTrust,
			ReleaseTrustRoot: releaseRoot,
		},
	)
	if err != nil {
		return nil, false, err
	}
	claims, err := tenancy.NewCustomerWorkerStore(pool)
	if err != nil {
		return nil, false, err
	}
	onboarding, err := tenancy.NewCustomerOnboardingCoordinator(pool)
	if err != nil {
		return nil, false, err
	}
	recorder, err := tenancy.NewCustomerAuthorityRecorder(onboarding)
	if err != nil {
		return nil, false, err
	}
	directory, err := iamSettings.LoginDirectoryKeyring()
	if err != nil {
		return nil, false, err
	}
	activate, err := newCustomerActivationCallback(directory)
	if err != nil {
		return nil, false, err
	}
	finalizer, err := tenancy.NewCustomerActivationFinalizer(onboarding, activate)
	if err != nil {
		return nil, false, err
	}
	runner, err := tenancy.NewCustomerWorker(
		claims, acknowledger, recorder, finalizer,
		"harpy-v2:core-worker:customer-provisioning",
	)
	if err != nil {
		return nil, false, err
	}
	workload, err := newCustomerProvisioningWorkload(runner)
	return workload, err == nil, err
}
