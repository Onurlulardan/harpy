package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildInvitationHTTP(
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	settings config.IAMAPI,
	report httpedge.ErrorSink,
) error {
	application, err := iammodule.NewInvitationHTTPApplication(
		pool, coordinator, runtime.invitationRuntime, settings,
		runtime.deliverySealer, runtime.passwordBlocklists.Common(),
		runtime.passwordBlocklists.Compromised(),
	)
	if err != nil {
		return fmt.Errorf("build IAM invitation application: %w", err)
	}
	handler, err := runtime.security.NewInvitationHTTP(
		application, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM invitation HTTP: %w", err)
	}
	runtime.invitationHTTP = application
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
