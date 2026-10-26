package api

import (
	"fmt"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type deploymentHTTPFoundation struct {
	policy tenancymodule.DeploymentPolicy
	routes []httpedge.Route
}

func buildDeploymentHTTPFoundation(
	info buildinfo.Info,
	runtime *tenancymodule.DeploymentPolicyRuntime,
	report httpedge.ErrorSink,
) (deploymentHTTPFoundation, error) {
	if runtime == nil || report == nil {
		return deploymentHTTPFoundation{},
			fmt.Errorf("deployment HTTP dependencies are incomplete")
	}
	policy, err := runtime.Snapshot()
	if err != nil {
		return deploymentHTTPFoundation{},
			fmt.Errorf("read verified deployment policy: %w", err)
	}
	target, err := runtime.PrimaryTarget()
	if err != nil {
		return deploymentHTTPFoundation{},
			fmt.Errorf("read verified primary tenant target: %w", err)
	}
		publicSlug, _ := target.PublicTenantSlug()
	handler, err := newRuntimeConfigHandler(info, runtimeConfigSnapshot{
		deploymentMode:    config.DeploymentMode(policy.Mode()),
		primaryTenantSlug: publicSlug,
		aiChatComposed:    true,
		workComposed:      true,
	}, report)
	if err != nil {
		return deploymentHTTPFoundation{},
			fmt.Errorf("build runtime config HTTP: %w", err)
	}
	routes, err := runtimeConfigRoutes(handler)
	if err != nil {
		return deploymentHTTPFoundation{}, err
	}
	return deploymentHTTPFoundation{policy: policy, routes: routes}, nil
}
