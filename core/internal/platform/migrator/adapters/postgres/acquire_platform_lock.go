package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/acquire_platform_lock.sql
var acquirePlatformLockSQL string

func AcquirePlatformLock(
	ctx context.Context,
	connection *sql.Conn,
	lockKey string,
) error {
	if _, err := connection.ExecContext(ctx, acquirePlatformLockSQL, lockKey); err != nil {
		return fmt.Errorf("acquire platform migration advisory lock: %w", err)
	}
	return nil
}
