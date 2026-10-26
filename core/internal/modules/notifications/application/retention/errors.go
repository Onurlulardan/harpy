package retention

import "errors"

var (
	ErrClaimInvalid    = errors.New("notification retention claim is invalid")
	ErrConfiguration   = errors.New("notification retention configuration is invalid")
	ErrOutcomeInvalid  = errors.New("notification retention outcome is invalid")
	ErrReceiptConflict = errors.New("notification retention receipt conflicts")
)
