package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func RowNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
