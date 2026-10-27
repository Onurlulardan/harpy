//go:build integration

package executor

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type pg18Bound struct {
	database postgres.BoundExecutor
}

type pg18OwnerPort struct {
	owner       string
	binder      postgres.OwnerBinder[pg18Bound]
	observation *pg18Observation
	calls       atomic.Int32
}

func pg18NewOwnerPort(t *testing.T, owner string) *pg18OwnerPort {
	t.Helper()
	binder, err := postgres.NewOwnerBinder(
		owner, func(database postgres.BoundExecutor) pg18Bound {
			return pg18Bound{database: database}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &pg18OwnerPort{
		owner: owner, binder: binder, observation: &pg18Observation{},
	}
}

func (port *pg18OwnerPort) Run(
	scope transaction.TenantScope,
	program execution.Program,
) error {
	port.calls.Add(1)
	bound, err := transaction.Bind(scope, port.binder)
	if err != nil {
		return err
	}
	driver := &pg18Driver{
		owner: port.owner, scope: scope, database: bound.database,
		observation: port.observation,
	}
	return program.Run(driver)
}

func (port *pg18OwnerPort) bind(
	scope transaction.TenantScope,
) (pg18Bound, error) {
	return transaction.Bind(scope, port.binder)
}

type pg18Driver struct {
	owner       string
	scope       transaction.TenantScope
	database    postgres.BoundExecutor
	observation *pg18Observation
}

func (driver *pg18Driver) Owner() string                  { return driver.owner }
func (driver *pg18Driver) Scope() transaction.TenantScope { return driver.scope }

func (driver *pg18Driver) Execute(
	ctx context.Context, statement string, arguments ...any,
) error {
	driver.observation.record("execute", statement, arguments)
	_, err := driver.database.Execute(ctx, statement, arguments...)
	driver.observation.recordDatabaseError(err)
	return err
}

func (driver *pg18Driver) Query(
	ctx context.Context, statement string, arguments ...any,
) (execution.Rows, error) {
	driver.observation.record("query", statement, arguments)
	rows, err := driver.database.Query(ctx, statement, arguments...)
	if err != nil {
		driver.observation.recordDatabaseError(err)
		return nil, err
	}
	driver.observation.opened.Add(1)
	return &pg18Rows{Rows: rows, observation: driver.observation}, nil
}

func (driver *pg18Driver) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) execution.Row {
	driver.observation.record("row", statement, arguments)
	return &pg18Row{
		Row:         driver.database.QueryRow(ctx, statement, arguments...),
		observation: driver.observation,
	}
}

type pg18Row struct {
	postgres.Row
	observation *pg18Observation
}

func (row *pg18Row) Scan(destinations ...any) error {
	err := row.Row.Scan(destinations...)
	row.observation.recordDatabaseError(err)
	return err
}

type pg18Rows struct {
	postgres.Rows
	observation *pg18Observation
	once        sync.Once
}

func (rows *pg18Rows) Close() {
	rows.once.Do(func() {
		rows.Rows.Close()
		rows.observation.closeRows()
	})
}

func (rows *pg18Rows) Scan(destinations ...any) error {
	err := rows.Rows.Scan(destinations...)
	rows.observation.recordDatabaseError(err)
	return err
}

func (rows *pg18Rows) Err() error {
	err := rows.Rows.Err()
	rows.observation.recordDatabaseError(err)
	return err
}
