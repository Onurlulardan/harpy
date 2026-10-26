package postgres

import (
	"errors"
	"fmt"
)

var (
	ErrDownForbidden             = errors.New("migration down is forbidden")
	ErrCreateTransactionRequired = errors.New("history creation requires an active transaction")
)

type DownForbiddenError struct {
	Version int64
}

func (failure *DownForbiddenError) Error() string {
	return fmt.Sprintf("migration %d: %v", failure.Version, ErrDownForbidden)
}

func (failure *DownForbiddenError) Unwrap() error {
	return ErrDownForbidden
}
