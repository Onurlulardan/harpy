package pgtools

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
)

func (tools *ArchiveTools) Dump(
	ctx context.Context,
	dsn string,
	snapshotID string,
	target io.Writer,
) error {
	cleanDSN, password, err := sanitizedDSN(dsn)
	if tools == nil || err != nil || ctx == nil || ctx.Err() != nil || target == nil ||
		!validSnapshot(snapshotID) {
		return errors.New("PostgreSQL backup request is invalid")
	}
	var stderr bytes.Buffer
	err = tools.runner.Run(ctx, Command{Executable: tools.pgDump,
		Arguments: []string{"--format=custom", "--no-owner",
			"--snapshot=" + snapshotID, "--dbname=" + cleanDSN},
		Environment: []string{"PGPASSWORD=" + password}, Stdout: target, Stderr: &stderr})
	if err != nil {
		return errors.New("PostgreSQL backup command failed")
	}
	return nil
}

func validSnapshot(value string) bool {
	return len(value) >= 1 && len(value) <= 256 &&
		!strings.ContainsAny(value, "\x00\r\n")
}
