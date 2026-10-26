package v1cutover

import "errors"

var (
	ErrInvalid     = errors.New("notification cutover input is invalid")
	ErrConflict    = errors.New("notification cutover state conflicts")
	ErrNotReady    = errors.New("notification cutover evidence is not ready")
	ErrUnavailable = errors.New("notification cutover storage is unavailable")
)
