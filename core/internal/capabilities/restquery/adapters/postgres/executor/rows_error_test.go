package executor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestRowsTerminalErrorClosesAndDoesNotLeak(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	record := fixtureRecordRow(
		t, batch, "Ada", nil,
		"40000000-0000-4000-8000-000000000001",
	)
	driver, statements := readyDriver(t, batch, [][]any{record}, 0)
	rows := driver.rows[statements[0].sql]
	rows.err = errors.New("SECRET terminal driver detail")

	_, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if !errors.Is(err, ErrExecution) || strings.Contains(err.Error(), "SECRET") {
		t.Fatalf("terminal error=%v", err)
	}
	if !rows.closed {
		t.Fatal("rows were not closed after terminal error")
	}
}

func TestRowShapeErrorClosesWithoutPartialPage(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	driver, statements := readyDriver(t, batch, [][]any{{"wrong-width"}}, 0)
	rows := driver.rows[statements[0].sql]

	result, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if !errors.Is(err, ErrContract) || result.Validate() == nil {
		t.Fatalf("shape error=%v result valid=%v", err, result.Validate() == nil)
	}
	if !rows.closed {
		t.Fatal("rows were not closed after shape error")
	}
}

func TestRecordProbeOverflowClosesRows(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	values := make([][]any, 4)
	for index := range values {
		values[index] = fixtureRecordRow(
			t, batch, "Ada", nil,
			"40000000-0000-4000-8000-000000000001",
		)
	}
	driver, statements := readyDriver(t, batch, values, 0)
	_, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("overflow error=%v", err)
	}
	if !driver.rows[statements[0].sql].closed {
		t.Fatal("overflow rows were not closed")
	}
}
