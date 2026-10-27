package executor

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestRunBatchExplainsAllStatementsBeforeData(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeExact, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	record := fixtureRecordRow(
		t, batch, "Ada", nil,
		"40000000-0000-4000-8000-000000000001",
	)
	driver, statements := readyDriver(t, batch, [][]any{record}, 7)

	result, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatalf("run: %v events=%v", err, driver.events)
	}
	assertExecutionOrder(t, driver, statements)
	page, pageErr := result.Page()
	records, recordsErr := page.Records()
	total, present, countErr := result.ExactCount()
	if pageErr != nil || recordsErr != nil || countErr != nil ||
		len(records) != 1 || string(records[0]) !=
		`{"name":"Ada","metadata":null}` || total != 7 || !present {
		t.Fatalf("records=%q total=%d present=%v errors=%v/%v/%v",
			records, total, present, pageErr, recordsErr, countErr)
	}
	if !driver.rows[statements[0].sql].closed {
		t.Fatal("record rows were not closed")
	}
}

func assertExecutionOrder(
	t *testing.T,
	driver *fakeDriver,
	statements []statement,
) {
	t.Helper()
	wantKinds := []string{
		"execute", "row", "execute", "row",
		"execute", "query", "execute", "row",
	}
	if len(driver.calls) != len(wantKinds) {
		t.Fatalf("calls=%d events=%v", len(driver.calls), driver.events)
	}
	for index, kind := range wantKinds {
		if driver.calls[index].kind != kind {
			t.Fatalf("call %d kind=%s events=%v", index, driver.calls[index].kind, driver.events)
		}
	}
	if driver.calls[1].statement != statements[0].explainSQL ||
		driver.calls[3].statement != statements[1].explainSQL ||
		driver.calls[5].statement != statements[0].sql ||
		driver.calls[7].statement != statements[1].sql {
		t.Fatalf("EXPLAIN/data ordering differs: %v", driver.events)
	}
	if strings.Contains(statements[0].explainSQL, "ANALYZE") ||
		strings.Contains(statements[1].explainSQL, "ANALYZE") {
		t.Fatal("EXPLAIN ANALYZE was used")
	}
	if !reflect.DeepEqual(driver.calls[1].arguments, statements[0].arguments) ||
		!reflect.DeepEqual(driver.calls[5].arguments, statements[0].arguments) ||
		!reflect.DeepEqual(driver.calls[3].arguments, statements[1].arguments) ||
		!reflect.DeepEqual(driver.calls[7].arguments, statements[1].arguments) {
		t.Fatal("EXPLAIN and data did not use exact arguments")
	}
	if !reflect.DeepEqual(driver.calls[0].arguments, []any{"5000ms"}) ||
		!reflect.DeepEqual(driver.calls[2].arguments, []any{"2000ms"}) ||
		!reflect.DeepEqual(driver.calls[4].arguments, []any{"5000ms"}) ||
		!reflect.DeepEqual(driver.calls[6].arguments, []any{"2000ms"}) {
		t.Fatal("class-local timeouts differ")
	}
}
