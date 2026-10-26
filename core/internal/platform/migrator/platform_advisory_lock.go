package migrator

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

type platformAdvisoryLock struct {
	connection *sql.Conn
}

func (runner *Runner) acquirePlatformLock(
	ctx context.Context,
) (*platformAdvisoryLock, error) {
	if runner.database.Stats().MaxOpenConnections == 1 {
		return nil, dependencyPreflightError(
			fmt.Errorf("platform lock requires a second migration connection"),
		)
	}
	connection, err := runner.database.Conn(ctx)
	if err != nil {
		return nil, dependencyPreflightError(
			fmt.Errorf("acquire migration lock connection: %w", err),
		)
	}
	if _, err := runner.preflightDatabase(ctx, connection); err != nil {
		_ = connection.Close()
		return nil, dependencyPreflightError(err)
	}
	if err := runner.preflightExtensions(ctx, connection); err != nil {
		_ = connection.Close()
		return nil, dependencyPreflightError(err)
	}
	if err := postgresstore.AcquirePlatformLock(
		ctx, connection, platformMigrationLockKey,
	); err != nil {
		discardConnection(connection)
		return nil, dependencyPreflightError(
			fmt.Errorf("acquire platform migration advisory lock: %w", err),
		)
	}
	return &platformAdvisoryLock{connection: connection}, nil
}

func (runner *Runner) releasePlatformLock(
	ctx context.Context,
	lock *platformAdvisoryLock,
) error {
	if lock == nil || lock.connection == nil {
		return nil
	}
	cleanupContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), runner.cleanupTimeout,
	)
	defer cancel()
	err := postgresstore.ReleasePlatformLock(
		cleanupContext, lock.connection, platformMigrationLockKey,
	)
	if err != nil {
		discardConnection(lock.connection)
		return err
	}
	if err := lock.connection.Close(); err != nil {
		return fmt.Errorf("close migration lock connection: %w", err)
	}
	return nil
}

func discardConnection(connection *sql.Conn) {
	if connection == nil {
		return
	}
	_ = connection.Raw(func(any) error { return driver.ErrBadConn })
	_ = connection.Close()
}
