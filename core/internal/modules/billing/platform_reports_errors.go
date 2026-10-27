package billing

import (
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func normalizePlatformReportError(err error) error {
	if _, _, _, valid := billingapi.QueryErrorDetails(err); valid {
		return err
	}
	if errors.Is(err, billingapi.ErrForbidden) || errors.Is(err, billingapi.ErrAuthorityUnavailable) {
		return err
	}
	return billingapi.ErrUnavailable
}

func normalizePlatformTransitionError(err error) error {
	switch {
	case errors.Is(err, platformreportapp.ErrTarget):
		return billingapi.ErrNotFound
	case errors.Is(err, subscriptionapp.ErrRequest):
		return billingapi.ErrInvalid
	case errors.Is(err, subscriptionapp.ErrStaleRevision):
		return billingapi.ErrRevisionStale
	case errors.Is(err, subscriptionapp.ErrIdempotencyReused):
		return billingapi.ErrIdempotencyReused
	case errors.Is(err, subscriptionapp.ErrStateConflict):
		return billingapi.ErrConflict
	case errors.Is(err, billingapi.ErrForbidden),
		errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return err
	default:
		return billingapi.ErrUnavailable
	}
}
