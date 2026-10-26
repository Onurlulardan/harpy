package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	workmodule "github.com/Onurlulardan/harpy/core/internal/modules/work"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type workHTTPRuntime struct {
	handler     *workmodule.HTTPHandler
	application *workmodule.PlatformHTTPRuntime
}

func buildWorkHTTP(settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, iamAccess workIAMAccess,
	report httpedge.ErrorSink,
) (workHTTPRuntime, error) {
	if pool == nil || iamAccess == nil || report == nil {
		return workHTTPRuntime{}, fmt.Errorf("Work HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return workHTTPRuntime{}, fmt.Errorf("build Work coordinator: %w", err)
	}
	access, err := newWorkAccessVerifier(iamAccess)
	if err != nil {
		return workHTTPRuntime{}, err
	}
	principals, err := iammodule.NewWorkPrincipalRuntime()
	if err != nil {
		return workHTTPRuntime{}, fmt.Errorf("build Work IAM principal runtime: %w", err)
	}
	resources, err := buildWorkHTTPResources(settings)
	if err != nil {
		return workHTTPRuntime{}, err
	}
	resources.Generations = pool
	resources.Scopes = coordinator
	resources.Catalogs = catalogs
	resources.Principals = principals
	application, err := workmodule.NewPlatformHTTPRuntime(resources)
	if err != nil {
		return workHTTPRuntime{}, fmt.Errorf("build Work HTTP application: %w", err)
	}
	handler, err := workmodule.NewHTTPHandler(workmodule.HTTPHandlerDependencies{
		Access: access, Application: application, Report: workmodule.HTTPErrorSink(report)})
	if err != nil || !handler.Ready() {
		return workHTTPRuntime{}, fmt.Errorf("build Work HTTP handler: %w", err)
	}
	return workHTTPRuntime{handler: handler, application: application}, nil
}

func (runtime workHTTPRuntime) Ready() bool {
	return runtime.handler != nil && runtime.handler.Ready() &&
		runtime.application != nil && len(runtime.handler.Routes()) == 68
}

func (runtime workHTTPRuntime) Routes() []httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return runtime.handler.Routes()
}
