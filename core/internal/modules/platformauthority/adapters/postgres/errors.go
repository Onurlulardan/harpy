package postgres

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type sqlStateError interface{ SQLState() string }

func (store *Store) mapDatabaseError(err error) error {
	if store != nil && store.rowNotFound != nil && store.rowNotFound(err) {
		return ports.ErrNotFound
	}
	var postgresError sqlStateError
	if errors.As(err, &postgresError) {
		switch postgresError.SQLState() {
		case "40001":
			return ports.ErrStaleHead
		case "23505", "23514", "23503", "55000":
			return ports.ErrConflict
		case "P0002":
			return ports.ErrNotFound
		}
	}
	return err
}
