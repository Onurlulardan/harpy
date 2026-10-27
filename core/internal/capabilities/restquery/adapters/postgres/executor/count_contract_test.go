package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestExactCountCannotBeOmittedOrNegative(t *testing.T) {
	tests := []struct {
		name  string
		value any
		err   error
	}{
		{name: "missing", err: errors.New("SECRET count detail")},
		{name: "negative", value: int64(-1)},
		{name: "wrong shape", value: "7"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newBatchFixture(
				t, api.CountModeExact, fixtureTenant, fixtureActor,
			)
			batch := consumedFixture(t, fixture)
			driver, statements := readyDriver(t, batch, nil, 0)
			driver.rowValues[statements[1].sql] = test.value
			driver.rowErrors[statements[1].sql] = test.err

			result, err := runBatch(
				context.Background(), driver, batch,
				fixture.codec, planner.HardPolicy(),
			)
			if !errors.Is(err, ErrContract) || result.Validate() == nil {
				t.Fatalf("count error=%v result valid=%v", err, result.Validate() == nil)
			}
		})
	}
}

func TestNoneCountModeNeverExecutesCount(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	driver, statements := readyDriver(t, batch, nil, 0)
	result, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	total, present, countErr := result.ExactCount()
	if countErr != nil || total != 0 || present {
		t.Fatalf("count=%d present=%v error=%v", total, present, countErr)
	}
	if len(statements) != 1 || len(driver.calls) != 4 {
		t.Fatalf("none count calls=%v", driver.events)
	}
}
