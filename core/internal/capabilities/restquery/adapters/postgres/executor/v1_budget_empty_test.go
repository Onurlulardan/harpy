package executor

import (
	"context"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestRunV1RejectsPageStageBeforePageAndCountData(t *testing.T) {
	fixture := newV1RunFixture(t)
	driver := fixture.driver(t, "50000", "50000.1")
	_, err := runV1(
		context.Background(), driver, fixture.plan, fixture.ordinal,
		fixture.predecessor, fixture.batch, fixture.compiler,
		fixture.codec, planner.HardPolicy(),
	)
	assertQueryCode(t, err, restqueryv1.ErrorQueryBudgetExceeded)
	if eventIndex(driver.events, "query:"+fixture.seek.sql) < 0 ||
		eventIndex(driver.events, "row:"+fixture.pageSQL.explainSQL) < 0 {
		t.Fatal("predecessor/page admission stage did not run")
	}
	if eventIndex(driver.events, "query:"+fixture.pageSQL.sql) >= 0 ||
		eventIndex(driver.events, "row:"+fixture.count.sql) >= 0 {
		t.Fatalf("data ran after page budget rejection: %v", driver.events)
	}
}

func TestRunV1MissingPredecessorReturnsCanonicalEmptyWithExactCount(t *testing.T) {
	fixture := newV1RunFixture(t)
	driver := fixture.driver(t, "1", "1")
	driver.rows[fixture.seek.sql] = &fakeRows{}
	result, err := runV1(
		context.Background(), driver, fixture.plan, fixture.ordinal,
		fixture.predecessor, fixture.batch, fixture.compiler,
		fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := result.Page()
	records, _ := page.Records()
	total, present, _ := result.ExactCount()
	if len(records) != 0 || total != 3 || !present {
		t.Fatalf("records=%q total=%d present=%t", records, total, present)
	}
	if eventIndex(driver.events, "row:"+fixture.pageSQL.explainSQL) >= 0 ||
		eventIndex(driver.events, "query:"+fixture.pageSQL.sql) >= 0 {
		t.Fatalf("missing predecessor fell back to a page: %v", driver.events)
	}
}
