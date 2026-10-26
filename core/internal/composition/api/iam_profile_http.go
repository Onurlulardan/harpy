package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildProfileHTTP(
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	changes iammodule.IdentityMutationSink,
	report httpedge.ErrorSink,
) error {
	application, err := iammodule.NewProfileHTTPApplication(
		pool, coordinator, runtime.profileCommands, runtime.profileBackfills,
		runtime.profileRecoveries, changes,
	)
	if err != nil {
		return fmt.Errorf("build IAM profile application: %w", err)
	}
	handler, err := runtime.security.NewProfileHTTP(
		application, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM profile HTTP: %w", err)
	}
	runtime.profileHTTP = application
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
