package executionpostgres

import (
	"context"

	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
)

func (*driver) Owner() string { return offering.Owner }

func (driver *driver) GlobalScope() globalScope {
	if driver == nil {
		return globalScope{}
	}
	return driver.scope
}

func (driver *driver) Execute(
	ctx context.Context,
	statement string,
	arguments ...any,
) error {
	if driver == nil {
		return offering.ErrContract
	}
	_, err := driver.database.Execute(ctx, statement, arguments...)
	return err
}

func (driver *driver) Query(
	ctx context.Context,
	statement string,
	arguments ...any,
) (restexecution.Rows, error) {
	if driver == nil {
		return nil, offering.ErrContract
	}
	return driver.database.Query(ctx, statement, arguments...)
}

func (driver *driver) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) restexecution.Row {
	if driver == nil {
		return failedRow{}
	}
	return driver.database.QueryRow(ctx, statement, arguments...)
}

type failedRow struct{}

func (failedRow) Scan(...any) error { return offering.ErrContract }
