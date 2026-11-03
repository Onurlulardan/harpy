package goose

import (
	"context"
	"database/sql"
	"io/fs"

	pressly "github.com/pressly/goose/v3"
)

func upDatabase(ctx context.Context, database *sql.DB, source fs.FS) error {
	provider, err := pressly.NewProvider(pressly.DialectPostgres, database, source)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}
