package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildLifecycleHTTP(pool *postgres.Pool,
	coordinator *platformtx.Coordinator, settings config.IAMAPI,
	report httpedge.ErrorSink) error {
	application, err := iammodule.NewLifecycleHTTPApplication(pool, coordinator,
		runtime.lifecycleRuntime, settings, runtime.deliverySealer, runtime.totpSealer,
		runtime.totpDecryptor, runtime.passwordBlocklists.Common(),
		runtime.passwordBlocklists.Compromised())
	if err != nil {
		return fmt.Errorf("build IAM lifecycle application: %w", err)
	}
	handler, err := runtime.security.NewLifecycleHTTP(application, iammodule.HTTPErrorSink(report))
	if err != nil {
		return fmt.Errorf("build IAM lifecycle HTTP: %w", err)
	}
	runtime.lifecycleHTTP = application
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
