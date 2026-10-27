package retention

import "errors"

var (
	ErrClaimInvalid   = errors.New("Analytics retention claim is invalid")
	ErrConfiguration  = errors.New("Analytics retention configuration is invalid")
	ErrOutcomeInvalid = errors.New("Analytics retention outcome is invalid")
)
