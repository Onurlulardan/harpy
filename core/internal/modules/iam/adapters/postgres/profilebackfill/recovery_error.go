package profilebackfill

import (
	"errors"
	"fmt"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

type recoverySQLStateError interface{ SQLState() string }

func recoveryStoreError(err error) error {
	var state recoverySQLStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "40001", "P5742":
			return profilebackfillapp.ErrConflict
		}
	}
	return fmt.Errorf("%w: %w", profilebackfillapp.ErrUnavailable, err)
}
