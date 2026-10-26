package executionpostgres

import (
	"context"

	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
)

func (*driver) Owner() string { return workquery.Owner }

func (driver *driver) Scope() tenantScope {
	if driver == nil {
		return tenantScope{}
	}
	return driver.scope
}

func (driver *driver) Execute(ctx context.Context, statement string,
	arguments ...any,
) error {
	if driver == nil {
		return workquery.ErrUnavailable
	}
	_, err := driver.database.Execute(ctx, statement, arguments...)
	return err
}

func (driver *driver) Query(ctx context.Context, statement string,
	arguments ...any,
) (restexecution.Rows, error) {
	if driver == nil {
		return nil, workquery.ErrUnavailable
	}
	return driver.database.Query(ctx, statement, arguments...)
}

func (driver *driver) QueryRow(ctx context.Context, statement string,
	arguments ...any,
) restexecution.Row {
	if driver == nil {
		return failedRow{}
	}
	return driver.database.QueryRow(ctx, statement, arguments...)
}

type failedRow struct{}

func (failedRow) Scan(...any) error { return workquery.ErrUnavailable }
