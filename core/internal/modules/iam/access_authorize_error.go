package iam

import (
	"errors"
	"fmt"
)

func mapAccessAuthorizeError(forbidden, unavailable, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, forbidden) {
		return forbidden
	}
	if errors.Is(err, unavailable) {
		return err
	}
	return fmt.Errorf("%w: %w", unavailable, err)
}
