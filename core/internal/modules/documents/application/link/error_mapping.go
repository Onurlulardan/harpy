package link

import "errors"

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrForbidden),
		errors.Is(err, ErrNotFound), errors.Is(err, ErrConflict):
		return err
	default:
		return ErrUnavailable
	}
}
