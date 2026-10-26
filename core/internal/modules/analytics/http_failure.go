package analytics

import (
	"errors"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func analyticsHTTPFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, queryapp.ErrEntitlement),
		errors.Is(err, dashboardapp.ErrEntitlement),
		errors.Is(err, publicationapp.ErrEntitlement):
		return 403, "ANALYTICS_ENTITLEMENT_DENIED", "Analytics entitlement denied the operation"
	case errors.Is(err, queryapp.ErrForbidden), errors.Is(err, dashboardapp.ErrForbidden),
		errors.Is(err, publicationapp.ErrForbidden), errors.Is(err, exportapp.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, queryapp.ErrNotFound), errors.Is(err, dashboardapp.ErrNotFound),
		errors.Is(err, publicationapp.ErrNotFound), errors.Is(err, compatibilityapp.ErrNotFound):
		return 404, "ANALYTICS_NOT_FOUND", "Analytics resource was not found"
	case errors.Is(err, exportapp.ErrNotFound):
		return 404, "ANALYTICS_EXPORT_NOT_FOUND", "Dashboard export was not found"
	case errors.Is(err, exportapp.ErrIdempotencyReused):
		return 409, "IDEMPOTENCY_KEY_REUSED", "Idempotency key was already used"
	case errors.Is(err, exportapp.ErrNotReady):
		return 409, "ANALYTICS_EXPORT_NOT_READY", "Dashboard export is not ready"
	case errors.Is(err, exportapp.ErrQueueLimit):
		return 429, "ANALYTICS_EXPORT_QUEUE_LIMIT_EXCEEDED", "Dashboard export queue limit was exceeded"
	case errors.Is(err, exportapp.ErrExpired):
		return 410, "ANALYTICS_EXPORT_EXPIRED", "Dashboard export expired"
	case errors.Is(err, queryapp.ErrConflict), errors.Is(err, dashboardapp.ErrConflict),
		errors.Is(err, publicationapp.ErrConflict), errors.Is(err, compatibilityapp.ErrConflict),
		errors.Is(err, compatibilityapp.ErrBlocked), errors.Is(err, exportapp.ErrConflict):
		return 409, "ANALYTICS_CONFLICT", "Analytics state conflicts"
	case errors.Is(err, publicationapp.ErrRateLimited):
		return 429, "ANALYTICS_RATE_LIMITED", "Public request rate limit was exceeded"
	case errors.Is(err, queryapp.ErrInvalid), errors.Is(err, dashboardapp.ErrInvalid),
		errors.Is(err, publicationapp.ErrInvalid), errors.Is(err, compatibilityapp.ErrInvalid),
		errors.Is(err, exportapp.ErrInvalid):
		return 422, "ANALYTICS_INPUT_INVALID", "Analytics input is invalid"
	default:
		return 503, "ANALYTICS_UNAVAILABLE", "Analytics is unavailable"
	}
}
