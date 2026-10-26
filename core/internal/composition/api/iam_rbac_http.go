package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildRBACHTTP(
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	report httpedge.ErrorSink,
) error {
	application, err := iammodule.NewRBACHTTPApplication(
		pool, coordinator, runtime.rbacRuntime,
	)
	if err != nil {
		return fmt.Errorf("build IAM RBAC application: %w", err)
	}
	handler, err := runtime.security.NewRBACHTTP(
		application, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM RBAC HTTP: %w", err)
	}
	runtime.rbacHTTP = application
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
