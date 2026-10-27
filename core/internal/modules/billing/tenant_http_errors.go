package billing

import (
	"errors"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func normalizeSubscriptionQueryError(err error) error {
	switch {
	case errors.Is(err, subscriptionapp.ErrCurrentSubscriptionNotFound):
		return billingapi.ErrNotFound
	case errors.Is(err, subscriptionapp.ErrSubscriptionQueryDependency),
		errors.Is(err, subscriptionapp.ErrQueryScope):
		return billingapi.ErrAuthorityUnavailable
	case errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return billingapi.ErrAuthorityUnavailable
	default:
		return billingapi.ErrUnavailable
	}
}

func normalizeUsageQueryError(err error) error {
	if errors.Is(err, usageapp.ErrQueryDependency) || errors.Is(err, usageapp.ErrQueryScope) ||
		errors.Is(err, billingapi.ErrAuthorityUnavailable) {
		return billingapi.ErrAuthorityUnavailable
	}
	return billingapi.ErrUnavailable
}

func normalizeDiscountQueryError(err error) error {
	switch {
	case errors.Is(err, discountapp.ErrEligibilityMissing):
		return billingapi.ErrNotFound
	case errors.Is(err, discountapp.ErrQueryDependency),
		errors.Is(err, discountapp.ErrQueryScope),
		errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return billingapi.ErrAuthorityUnavailable
	default:
		return billingapi.ErrUnavailable
	}
}

func normalizeSubscriptionCommandError(err error) error {
	switch {
	case errors.Is(err, subscriptionapp.ErrRequest):
		return billingapi.ErrInvalid
	case errors.Is(err, subscriptionapp.ErrStaleRevision):
		return billingapi.ErrRevisionStale
	case errors.Is(err, subscriptionapp.ErrIdempotencyReused):
		return billingapi.ErrIdempotencyReused
	case errors.Is(err, subscriptionapp.ErrStateConflict),
		errors.Is(err, subscriptionapp.ErrCancellationMissing),
		errors.Is(err, subscriptionapp.ErrCancellationExists),
		errors.Is(err, subscriptionapp.ErrChangeExists):
		return billingapi.ErrConflict
	case errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return billingapi.ErrAuthorityUnavailable
	default:
		return billingapi.ErrUnavailable
	}
}

func normalizeDiscountCommandError(err error) error {
	switch {
	case errors.Is(err, discountapp.ErrRequest):
		return billingapi.ErrInvalid
	case errors.Is(err, discountapp.ErrIneligible):
		return billingapi.ErrConflict
	case errors.Is(err, discountapp.ErrOperationReused):
		return billingapi.ErrIdempotencyReused
	case errors.Is(err, billingapi.ErrAuthorityUnavailable):
		return billingapi.ErrAuthorityUnavailable
	default:
		return billingapi.ErrUnavailable
	}
}
