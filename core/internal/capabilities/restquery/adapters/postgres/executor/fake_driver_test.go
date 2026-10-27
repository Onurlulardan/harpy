package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type fakeDriver struct {
	owner       string
	rows        map[string]*fakeRows
	rowValues   map[string]any
	rowErrors   map[string]error
	queryErrors map[string]error
	events      []string
	calls       []fakeCall
	onQuery     func(string)
	onCall      func(string, string)
}

type fakeCall struct {
	kind      string
	statement string
	arguments []any
}

func (driver *fakeDriver) Owner() string { return driver.owner }
func (*fakeDriver) Scope() transaction.TenantScope {
	return transaction.TenantScope{}
}
func (driver *fakeDriver) Execute(
	_ context.Context, statement string, arguments ...any,
) error {
	driver.events = append(driver.events, "execute:"+statement)
	driver.recordCall("execute", statement, arguments)
	return nil
}
func (driver *fakeDriver) Query(
	_ context.Context, statement string, arguments ...any,
) (execution.Rows, error) {
	driver.events = append(driver.events, "query:"+statement)
	driver.recordCall("query", statement, arguments)
	if driver.onQuery != nil {
		driver.onQuery(statement)
	}
	if err := driver.queryErrors[statement]; err != nil {
		return nil, err
	}
	return driver.rows[statement], nil
}
func (driver *fakeDriver) QueryRow(
	_ context.Context, statement string, arguments ...any,
) execution.Row {
	driver.events = append(driver.events, "row:"+statement)
	driver.recordCall("row", statement, arguments)
	return &fakeRow{value: driver.rowValues[statement], err: driver.rowErrors[statement]}
}

func (driver *fakeDriver) recordCall(
	kind string, statement string, arguments []any,
) {
	if driver.onCall != nil {
		driver.onCall(kind, statement)
	}
	driver.calls = append(driver.calls, fakeCall{
		kind: kind, statement: statement, arguments: append([]any{}, arguments...),
	})
}

type fakeRow struct {
	value any
	err   error
}

func (row *fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return errors.New("unexpected fake row width")
	}
	return assignFakeValue(destinations[0], row.value)
}

type fakeRows struct {
	values  [][]any
	index   int
	err     error
	scanErr error
	closed  bool
}

func (rows *fakeRows) Close()     { rows.closed = true }
func (rows *fakeRows) Err() error { return rows.err }
func (rows *fakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}
func (rows *fakeRows) Scan(destinations ...any) error {
	if rows.scanErr != nil {
		return rows.scanErr
	}
	values := rows.values[rows.index-1]
	if len(values) != len(destinations) {
		return errors.New("unexpected fake rows width")
	}
	for index := range values {
		if err := assignFakeValue(destinations[index], values[index]); err != nil {
			return err
		}
	}
	return nil
}

func assignFakeValue(destination any, value any) error {
	switch target := destination.(type) {
	case *any:
		*target = value
	case *[]byte:
		if value == nil {
			*target = nil
			return nil
		}
		document, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("fake value is not bytes")
		}
		*target = bytes.Clone(document)
	case *int64:
		number, ok := value.(int64)
		if !ok {
			return fmt.Errorf("fake value is not int64")
		}
		*target = number
	case *int16:
		number, ok := value.(int16)
		if !ok {
			return fmt.Errorf("fake value is not int16")
		}
		*target = number
	case *uuid.UUID:
		identifier, ok := value.(uuid.UUID)
		if !ok {
			return fmt.Errorf("fake value is not UUID")
		}
		*target = identifier
	default:
		return fmt.Errorf("unsupported fake destination")
	}
	return nil
}

type fakeSQLStateError struct{ state string }

func (err fakeSQLStateError) Error() string    { return "driver detail must stay hidden" }
func (err fakeSQLStateError) SQLState() string { return err.state }
