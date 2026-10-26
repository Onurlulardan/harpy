package api

import (
	"fmt"

	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildSupportHTTP(
	settings config.API,
	pool *platformpostgres.Pool,
	access supportmodule.HTTPAccessVerifier,
	report httpedge.ErrorSink,
) (*supportmodule.HTTPHandler, error) {
	profile, err := newProfileQueryRuntime(pool)
	if err != nil {
		return nil, err
	}
	rateKeys, err := newSupportRateKeyResolver(
		newAnalyticsClientIPResolver(settings.AnalyticsTrustedProxies()),
	)
	if err != nil {
		return nil, err
	}
	handler, err := supportmodule.NewRuntimeHTTPHandler(
		supportmodule.RuntimeHTTPDependencies{
			Run: profile.supportRunner(), RowNotFound: platformpostgres.RowNotFound,
			Clock: moduleRuntimeClock{}, IDs: supportRuntimeIDs{},
			HTTP: supportmodule.HTTPHandlerDependencies{
				Access: access, RateKeys: rateKeys,
				Booking: supportmodule.HTTPBookingConfig{},
				Report:  supportmodule.HTTPErrorSink(report),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build Support HTTP: %w", err)
	}
	if !handler.Ready() || len(handler.Routes()) != 17 {
		return nil, fmt.Errorf("Support HTTP route contract is incomplete")
	}
	return handler, nil
}
