package formscleanup

import "errors"

var (
	ErrInvalid     = errors.New("documents Forms cleanup input is invalid")
	ErrConflict    = errors.New("documents Forms cleanup authority conflicts")
	ErrUnavailable = errors.New("documents Forms cleanup dependency is unavailable")
)

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrConflict):
		return err
	default:
		return ErrUnavailable
	}
}
