package deletion

import "errors"

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrConflict), errors.Is(err, ErrNotFound):
		return err
	default:
		return ErrUnavailable
	}
}
