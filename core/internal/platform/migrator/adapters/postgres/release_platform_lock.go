package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/release_platform_lock.sql
var releasePlatformLockSQL string

func ReleasePlatformLock(
	ctx context.Context,
	connection *sql.Conn,
	lockKey string,
) error {
	var unlocked bool
	if err := connection.QueryRowContext(ctx, releasePlatformLockSQL, lockKey).
		Scan(&unlocked); err != nil {
		return fmt.Errorf("release platform migration advisory lock: %w", err)
	}
	if !unlocked {
		return fmt.Errorf("release platform migration advisory lock: lock was not owned")
	}
	return nil
}
