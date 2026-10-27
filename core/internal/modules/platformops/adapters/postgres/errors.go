package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type sqlStateError interface{ SQLState() string }

func (store *Store) mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if store != nil && store.rowNotFound != nil && store.rowNotFound(err) {
		return ports.ErrNotFound
	}
	var postgresError sqlStateError
	if errors.As(err, &postgresError) {
		switch postgresError.SQLState() {
		case "23000", "23505", "23514", "23503", "40001", "42501", "55000":
			return ports.ErrConflict
		case "P0002":
			return ports.ErrNotFound
		}
	}
	return err
}
