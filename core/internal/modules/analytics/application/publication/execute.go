package publication

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *PublicApplication) Execute(
	ctx context.Context, request PublicRequest,
) (PublicResult, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || request.PublicID == [16]byte{} ||
		!request.ExpectedKind.Valid() || !slugPattern.MatchString(request.Slug) || len(request.Slug) > 100 ||
		net.ParseIP(request.TrustedClientIP) == nil || len(request.Parameters) > 64<<10 ||
		!json.Valid(request.Parameters) {
		return PublicResult{}, ErrInvalid
	}
	route, err := application.directory.Route(ctx, request.PublicID)
	if err != nil {
		return PublicResult{}, ErrNotFound
	}
	if !validPublicRoute(route, request.PublicID, application.clock.Now()) ||
		route.ResourceKind != request.ExpectedKind ||
		!validCapability(route, request.CapabilityToken, application.clock.Now()) {
		return PublicResult{}, ErrForbidden
	}
	lease, err := application.limiter.Admit(ctx, PublicAdmission{TenantID: route.TenantID,
		ReleaseID: route.ReleaseID, Generation: route.Generation,
		TrustedClientIP: request.TrustedClientIP})
	if err != nil || nilPort(lease) {
		if err == ErrRateLimited {
			return PublicResult{}, ErrRateLimited
		}
		return PublicResult{}, ErrRateLimitUnavailable
	}
	var result PublicResult
	executionErr := application.scopes.InPublicationSnapshot(ctx, route,
		func(scoped context.Context, scope transaction.TenantScope) error {
			value, executeErr := application.executor.Execute(scoped, scope, route, request.Parameters)
			if executeErr != nil {
				return executeErr
			}
			result = value
			return nil
		})
	releaseErr := lease.Close(ctx)
	if executionErr != nil || releaseErr != nil || !validPublicBody(result.Body) {
		return PublicResult{}, ErrUnavailable
	}
	return result, nil
}

func validPublicRoute(route PublicRoute, publicID [16]byte, now time.Time) bool {
	return route.PublicID == publicID && route.TenantID != uuid.Nil &&
		validUUIDv7(route.ReleaseID) && route.ResourceKind.Valid() &&
		validUUIDv7(route.ResourceVersionID) && route.State == analyticspublication.StateActive &&
		route.Generation > 0 && validTime(now)
}

func validCapability(route PublicRoute, token []byte, now time.Time) bool {
	switch route.Audience {
	case analyticspublication.AudiencePublic:
		return len(token) == 0
	case analyticspublication.AudienceCapability:
		if len(token) != 32 || route.CapabilityExpiresAt == nil ||
			!now.Before(*route.CapabilityExpiresAt) {
			return false
		}
		digest := sha256.Sum256(token)
		return subtle.ConstantTimeCompare(digest[:], route.CapabilitySHA256[:]) == 1
	default:
		return false
	}
}
