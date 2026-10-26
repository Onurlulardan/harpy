package subscription

import (
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

func PurchasedGrantActive(
	state State,
	window *billingperiod.Period,
	instant time.Time,
) (bool, error) {
	if !state.Valid() {
		return false, ErrState
	}
	if !state.purchasedGrantBearing() {
		if window != nil {
			return false, ErrGrant
		}
		return false, nil
	}
	if window == nil {
		return false, ErrGrant
	}
	return window.Contains(instant), nil
}
