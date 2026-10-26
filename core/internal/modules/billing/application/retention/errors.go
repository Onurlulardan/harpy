package retention

import "errors"

var (
	ErrClaimInvalid   = errors.New("billing retention claim is invalid")
	ErrConfiguration  = errors.New("billing retention configuration is invalid")
	ErrOutcomeInvalid = errors.New("billing retention outcome is invalid")
)
