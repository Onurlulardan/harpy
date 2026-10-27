package executor

import (
	"context"
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestSQLState57014MapsToClassSpecificTimeout(t *testing.T) {
	tests := []struct {
		name string
		want restqueryv1.ErrorCode
		fail func(*fakeDriver, []statement)
	}{
		{
			name: "record query", want: restqueryv1.ErrorQueryTimeout,
			fail: func(driver *fakeDriver, statements []statement) {
				driver.queryErrors[statements[0].sql] =
					fakeSQLStateError{state: queryCanceledSQLState}
			},
		},
		{
			name: "record scan", want: restqueryv1.ErrorQueryTimeout,
			fail: func(driver *fakeDriver, statements []statement) {
				driver.rows[statements[0].sql].scanErr =
					fakeSQLStateError{state: queryCanceledSQLState}
			},
		},
		{
			name: "exact count", want: restqueryv1.ErrorQueryCountTimeout,
			fail: func(driver *fakeDriver, statements []statement) {
				driver.rowErrors[statements[1].sql] =
					fakeSQLStateError{state: queryCanceledSQLState}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newBatchFixture(
				t, api.CountModeExact, fixtureTenant, fixtureActor,
			)
			batch := consumedFixture(t, fixture)
			record := fixtureRecordRow(
				t, batch, "Ada", nil,
				"40000000-0000-4000-8000-000000000001",
			)
			driver, statements := readyDriver(t, batch, [][]any{record}, 7)
			test.fail(driver, statements)
			_, err := runBatch(
				context.Background(), driver, batch,
				fixture.codec, planner.HardPolicy(),
			)
			assertQueryCode(t, err, test.want)
			if test.name == "record scan" && !driver.rows[statements[0].sql].closed {
				t.Fatal("rows were not closed after timeout scan")
			}
		})
	}
}

func TestContextCancellationIsNotMappedToTimeout(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeNone, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	driver, statements := readyDriver(t, batch, nil, 0)
	ctx, cancel := context.WithCancel(context.Background())
	driver.onQuery = func(statement string) {
		if statement == statements[0].sql {
			cancel()
		}
	}
	driver.queryErrors[statements[0].sql] =
		fakeSQLStateError{state: queryCanceledSQLState}

	_, err := runBatch(ctx, driver, batch, fixture.codec, planner.HardPolicy())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
	var queryError api.QueryError
	if errors.As(err, &queryError) {
		t.Fatal("context cancellation was converted into planner timeout")
	}
}
