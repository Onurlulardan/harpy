package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func (runtime *iamAPIRuntime) buildPortalUserHTTP(
	settings config.API, pool *postgres.Pool,
	resolver iammodule.PortalSubjectResolver, report httpedge.ErrorSink,
) error {
	if runtime == nil || pool == nil || resolver == nil || runtime.coordinator == nil ||
		runtime.lifecycleHTTP == nil || runtime.identityChanges == nil {
		return fmt.Errorf("IAM portal user HTTP dependencies are incomplete")
	}
	iamSettings, err := settings.IAM()
	if err != nil {
		return fmt.Errorf("load IAM portal user settings: %w", err)
	}
	application, err := iammodule.NewPortalUserHTTPApplication(
		pool, runtime.coordinator, runtime.identityRuntime, resolver, iamSettings,
		runtime.identityChanges, runtime.lifecycleHTTP,
	)
	if err != nil {
		return fmt.Errorf("build IAM portal user application: %w", err)
	}
	handler, err := runtime.security.NewPortalUserHTTP(
		application, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM portal user HTTP: %w", err)
	}
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
