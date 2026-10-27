package pgtools

import (
	"bytes"
	"context"
	"errors"
	"io"
)

func (tools *ArchiveTools) Restore(
	ctx context.Context,
	dsn string,
	archive io.Reader,
) error {
	cleanDSN, password, err := sanitizedDSN(dsn)
	if tools == nil || err != nil || ctx == nil || ctx.Err() != nil || archive == nil {
		return errors.New("PostgreSQL restore request is invalid")
	}
	var stderr bytes.Buffer
	err = tools.runner.Run(ctx, Command{Executable: tools.pgRestore,
		Arguments: []string{"--exit-on-error", "--single-transaction", "--clean",
			"--if-exists", "--no-owner",
			"--role=harpy_owner", "--dbname=" + cleanDSN},
		Environment: []string{"PGPASSWORD=" + password}, Stdin: archive, Stderr: &stderr})
	if err != nil {
		return errors.New("PostgreSQL restore command failed")
	}
	return nil
}
