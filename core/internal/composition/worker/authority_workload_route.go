package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type authorityWorkloadRoute string

const (
	authorityRouteLicenseRefresh authorityWorkloadRoute = "billing_license_refresh"
	authorityRouteSaaSRegistry   authorityWorkloadRoute = "platformauthority_registry"
)

func routeAuthorityWorkload(
	mode config.WorkerAuthorityMode,
) (authorityWorkloadRoute, error) {
	switch mode {
	case config.WorkerAuthorityConnected,
		config.WorkerAuthorityAirGap,
		config.WorkerAuthorityDevelopment:
		return authorityRouteLicenseRefresh, nil
	case config.WorkerAuthoritySaaS:
		return authorityRouteSaaSRegistry, nil
	default:
		return "", fmt.Errorf("worker authority route is invalid")
	}
}
