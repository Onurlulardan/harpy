package platformreport

import "errors"

var (
	ErrConfiguration = errors.New("billing platform report resolver is not configured")
	ErrTarget        = errors.New("billing platform subscription target is unavailable")
)
