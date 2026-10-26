package cursor

import "errors"

var (
	errInvalidCodec     = errors.New("REST Query cursor codec is invalid")
	errInvalidContext   = errors.New("REST Query cursor context is invalid")
	errInvalidSort      = errors.New("REST Query cursor sort is invalid")
	errInvalidLastValue = errors.New("REST Query cursor last value is invalid")
	errInvalidClaims    = errors.New("REST Query cursor claims are invalid")
	errEncodeFailed     = errors.New("REST Query cursor encoding failed")
)
