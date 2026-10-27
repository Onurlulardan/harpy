package folder

import "errors"

func mapStoreError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrConflict), errors.Is(err, ErrInvalid):
		return err
	default:
		return ErrUnavailable
	}
}
