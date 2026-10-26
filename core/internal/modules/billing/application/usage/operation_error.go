package usage

import (
	"errors"
	"fmt"
)

func usageOperationError(operation string, err error) error {
	for _, known := range []error{
		ErrConfiguration, ErrRequest, ErrSnapshot, ErrReservationReused,
		ErrOperationReused, ErrReservationAbsent, ErrState,
	} {
		if errors.Is(err, known) {
			return err
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
