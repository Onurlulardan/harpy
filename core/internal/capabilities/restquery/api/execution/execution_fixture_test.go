package execution

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type testContextKey struct{}

type testPort func(transaction.TenantScope, Program) error

func (port testPort) Run(scope transaction.TenantScope, program Program) error {
	return port(scope, program)
}

type testDriver struct {
	owner        string
	scope        transaction.TenantScope
	context      context.Context
	executeCalls atomic.Int32
	rows         *testRows
	row          *testRow
}

func (driver *testDriver) Owner() string                  { return driver.owner }
func (driver *testDriver) Scope() transaction.TenantScope { return driver.scope }

func (driver *testDriver) Execute(
	ctx context.Context,
	_ string,
	_ ...any,
) error {
	driver.context = ctx
	driver.executeCalls.Add(1)
	return nil
}

func (driver *testDriver) Query(
	ctx context.Context,
	_ string,
	_ ...any,
) (Rows, error) {
	driver.context = ctx
	if driver.rows == nil {
		driver.rows = &testRows{}
	}
	return driver.rows, nil
}

func (driver *testDriver) QueryRow(
	ctx context.Context,
	_ string,
	_ ...any,
) Row {
	driver.context = ctx
	if driver.row == nil {
		driver.row = &testRow{}
	}
	return driver.row
}

type testRows struct {
	closed       atomic.Bool
	closedSignal chan struct{}
	closeOnce    sync.Once
}

func (rows *testRows) Close() {
	rows.closed.Store(true)
	if rows.closedSignal != nil {
		rows.closeOnce.Do(func() { close(rows.closedSignal) })
	}
}
func (rows *testRows) Err() error                     { return nil }
func (rows *testRows) Next() bool                     { return !rows.closed.Load() }
func (rows *testRows) Scan(destinations ...any) error { return nil }

type testRow struct {
	scans atomic.Int32
}

func (row *testRow) Scan(...any) error {
	row.scans.Add(1)
	return nil
}

func testScopeIdentity() scopeIdentity {
	return scopeIdentity{
		tenantID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		actorID:  uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		mode:     transaction.SnapshotQuery, operationID: snapshotOperationID,
	}
}

func testScopeVerifier(context.Context, scopeIdentity) error { return nil }

func newTestProgram(
	t *testing.T,
	owner string,
	execute func(context.Context, SnapshotDriver) error,
	verify scopeVerifier,
) Program {
	t.Helper()
	ctx, cancel := context.WithCancel(
		context.WithValue(context.Background(), testContextKey{}, "captured"),
	)
	program, err := newProgram(
		ctx, cancel, testScopeIdentity(), owner, execute, verify,
	)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func newTestRegistry(t *testing.T, owner string, port OwnerPort) Registry {
	t.Helper()
	registration, err := NewRegistration(owner, port)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry([]Registration{registration})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
