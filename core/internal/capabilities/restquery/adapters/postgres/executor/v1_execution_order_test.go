package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestRunV1StagesExplainBeforeDataAndShareHundredThousandBudget(t *testing.T) {
	fixture := newV1RunFixture(t)
	driver := fixture.driver(t, "50000", "50000")
	result, err := runV1(
		context.Background(), driver, fixture.plan, fixture.ordinal,
		fixture.predecessor, fixture.batch, fixture.compiler,
		fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatalf("run V1: %v events=%v", err, driver.events)
	}
	documents, err := result.Page()
	if err != nil {
		t.Fatal(err)
	}
	records, _ := documents.Records()
	if len(records) != 1 || !strings.Contains(string(records[0]), `"name":"Cem"`) {
		t.Fatalf("records=%q", records)
	}
	assertV1ExecutionOrder(t, driver, fixture)
	if !driver.rows[fixture.seek.sql].closed || !driver.rows[fixture.pageSQL.sql].closed {
		t.Fatal("V1 rows were not closed")
	}
}

func assertV1ExecutionOrder(t *testing.T, driver *fakeDriver, fixture v1RunFixture) {
	t.Helper()
	want := []struct{ kind, sql string }{
		{"row", fixture.seek.explainSQL}, {"row", fixture.count.explainSQL},
		{"query", fixture.seek.sql}, {"row", fixture.pageSQL.explainSQL},
		{"query", fixture.pageSQL.sql}, {"row", fixture.count.sql},
	}
	position := -1
	for _, expected := range want {
		found := -1
		for index := position + 1; index < len(driver.calls); index++ {
			if driver.calls[index].kind == expected.kind &&
				driver.calls[index].statement == expected.sql {
				found = index
				break
			}
		}
		if found < 0 {
			t.Fatalf("missing ordered %s %s calls=%v", expected.kind, expected.sql, driver.events)
		}
		position = found
	}
}
