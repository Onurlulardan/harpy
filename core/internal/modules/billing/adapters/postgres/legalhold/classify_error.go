package legalhold

import (
	"errors"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
)

type stateError interface{ SQLState() string }

func classifyError(err error) error {
	var databaseError stateError
	if !errors.As(err, &databaseError) {
		return err
	}
	switch databaseError.SQLState() {
	case "23505", "P5730":
		return legalholdapp.ErrConflict
	case "P5731":
		return legalholdapp.ErrNotFound
	case "P5721":
		return legalholdapp.ErrDependency
	default:
		return err
	}
}
