package presentation

import (
	"errors"
	"fmt"
)

var ErrInvalid = errors.New("forms presentation is invalid")

func invalid(path, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalid, path, reason)
}
