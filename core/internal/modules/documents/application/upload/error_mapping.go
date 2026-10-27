package upload

import "errors"

func mapStoreError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrNotFound),
		errors.Is(err, ErrConflict), errors.Is(err, ErrInProgress):
		return err
	default:
		return ErrUnavailable
	}
}

func mapBillingError(err error) error {
	if errors.Is(err, ErrQuota) {
		return ErrQuota
	}
	return ErrUnavailable
}
