package licenseoffering

import "errors"

var (
	ErrConfiguration = errors.New("license offering publisher is misconfigured")
	ErrBinding       = errors.New("license offering binding is invalid")
	ErrProjection    = errors.New("license offering projection is invalid")
	ErrReceipt       = errors.New("license offering receipt is invalid")
)
