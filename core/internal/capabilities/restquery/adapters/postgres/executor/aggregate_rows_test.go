package executor

import (
	"context"
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestAggregateRowsCloseOnScanAndTerminalErrors(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	for name, configure := range map[string]func(*fakeRows){
		"scan": func(rows *fakeRows) {
			rows.values = [][]any{{"OPEN", int64(1), nil}}
			rows.scanErr = errors.New("sensitive scan detail")
		},
		"terminal": func(rows *fakeRows) {
			rows.err = errors.New("sensitive rows detail")
		},
	} {
		t.Run(name, func(t *testing.T) {
			driver, candidate := readyAggregateDriver(t, batch, nil, "1")
			configure(driver.rows[candidate.sql])
			_, err := runAggregate(
				context.Background(), driver, batch, planner.HardPolicy(),
			)
			if err == nil || !driver.rows[candidate.sql].closed {
				t.Fatalf("error=%v closed=%v", err, driver.rows[candidate.sql].closed)
			}
		})
	}
}

func TestAggregateMapsScanAndTerminalSQLStates(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	for name, test := range map[string]struct {
		scan     bool
		sqlState string
		code     restqueryv1.ErrorCode
	}{
		"scan overflow":     {true, "22003", restqueryv1.ErrorQueryAggregateOverflow},
		"terminal overflow": {false, "22003", restqueryv1.ErrorQueryAggregateOverflow},
		"scan timeout":      {true, "57014", restqueryv1.ErrorQueryTimeout},
		"terminal timeout":  {false, "57014", restqueryv1.ErrorQueryTimeout},
	} {
		t.Run(name, func(t *testing.T) {
			driver, candidate := readyAggregateDriver(t, batch, nil, "1")
			rows := driver.rows[candidate.sql]
			stateError := fakeSQLStateError{state: test.sqlState}
			if test.scan {
				rows.values = [][]any{{"OPEN", int64(1), nil}}
				rows.scanErr = stateError
			} else {
				rows.err = stateError
			}
			_, err := runAggregate(context.Background(), driver, batch, planner.HardPolicy())
			assertAggregateQueryCode(t, err, test.code)
			if !rows.closed {
				t.Fatal("rows not closed")
			}
		})
	}
}
