package iam

import (
	"context"
	"sort"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *authHTTPApplication) platformCapabilities(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims accesstoken.Claims,
	view sessionview.View,
) ([]string, error) {
	if claims.SecurityEpoch() != view.SecurityEpoch ||
		claims.PlatformAuthorizationClock() != view.PlatformAuthorizationClock {
		return nil, ErrAuthHTTPInvalidSession
	}
	if view.PlatformSubjectID == uuid.Nil {
		return []string{}, nil
	}
	if application.platformIAM == nil || view.PlatformAuthorizationClock < 1 {
		return nil, ErrAuthHTTPUnavailable
	}
	capabilities := make([]string, 0, len(iamapplication.PlatformCapabilityCatalog))
	for _, capability := range iamapplication.PlatformCapabilityCatalog {
		if !iamapplication.ModeAllowsPlatform(view.DeploymentMode, capability) {
			continue
		}
		decision, err := application.platformIAM.Authorize(
			ctx, scope, iamapplication.PlatformEvaluateQuery{
				ActorID: iamapplication.PlatformID(claims.Subject()), Capability: capability,
				ExpectedTokenVersion:         claims.TokenVersion(),
				ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
				ExpectedPlatformClock:        claims.PlatformAuthorizationClock(),
			},
		)
		if err != nil || decision.DeploymentMode != string(view.DeploymentMode) {
			return nil, ErrAuthHTTPUnavailable
		}
		if decision.Allowed {
			capabilities = append(capabilities, capability)
		}
	}
	sort.Strings(capabilities)
	return capabilities, nil
}
