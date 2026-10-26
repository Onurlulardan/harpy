package archive

import "errors"

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrForbidden),
		errors.Is(err, ErrNotFound), errors.Is(err, ErrConflict),
		errors.Is(err, ErrLimit), errors.Is(err, ErrCancelled):
		return err
	default:
		return ErrUnavailable
	}
}
