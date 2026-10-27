package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildPermissionRequestHTTP(
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	report httpedge.ErrorSink,
) error {
	application, err := iammodule.NewPermissionRequestHTTPApplication(
		pool, coordinator, runtime.permissionRequests,
	)
	if err != nil {
		return fmt.Errorf("build IAM permission request application: %w", err)
	}
	handler, err := runtime.security.NewPermissionRequestHTTP(
		application, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM permission request HTTP: %w", err)
	}
	runtime.permissionHTTP = application
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
