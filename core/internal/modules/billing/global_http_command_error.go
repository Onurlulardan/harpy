package billing

import (
	"errors"
	"strings"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

type sqlStateError interface{ SQLState() string }

func normalizeCommandError(err error) error {
	if errors.Is(err, errInvalidInput) || errors.Is(err, offeringapp.ErrRequest) ||
		errors.Is(err, discountapp.ErrRequest) || errors.Is(err, campaignapp.ErrRequest) {
		return billingapi.ErrInvalid
	}
	var databaseError sqlStateError
	if errors.As(err, &databaseError) {
		switch databaseError.SQLState() {
		case "40001":
			return billingapi.ErrRevisionStale
		case "23505":
			if strings.Contains(err.Error(), "idempotency key was reused") {
				return billingapi.ErrIdempotencyReused
			}
			return billingapi.ErrConflict
		case "23514", "55000":
			return billingapi.ErrConflict
		}
	}
	return billingapi.ErrAuthorityUnavailable
}

func normalizeCommandReadback[T any](value T, err error) (T, error) {
	if err != nil {
		return *new(T), billingapi.ErrAuthorityUnavailable
	}
	return value, nil
}
