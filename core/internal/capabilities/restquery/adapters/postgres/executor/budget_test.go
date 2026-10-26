package executor

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestBudgetRejectionRunsNoDataStatement(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeExact, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	driver, statements := readyDriver(t, batch, nil, 0)
	driver.rowValues[statements[0].explainSQL] = explainDocument("2")
	policy := fixtureLowPolicy(t)

	result, err := runBatch(
		context.Background(), driver, batch, fixture.codec, policy,
	)
	assertQueryCode(t, err, restqueryv1.ErrorQueryBudgetExceeded)
	if result.Validate() == nil {
		t.Fatal("budget rejection returned a partial result")
	}
	if eventIndex(driver.events, "row:"+statements[1].explainSQL) < 0 {
		t.Fatal("exact-count EXPLAIN was skipped before batch admission")
	}
	assertNoDataCalls(t, driver, statements)
}

func TestMalformedOrOversizedExplainRunsNoDataStatement(t *testing.T) {
	for _, document := range [][]byte{
		[]byte(`[{"Plan":{"Total Cost":1},"Plan":{"Total Cost":1}}]`),
		make([]byte, planner.MaximumExplainBytes+1),
	} {
		fixture := newBatchFixture(
			t, api.CountModeNone, fixtureTenant, fixtureActor,
		)
		batch := consumedFixture(t, fixture)
		driver, statements := readyDriver(t, batch, nil, 0)
		driver.rowValues[statements[0].explainSQL] = document
		if _, err := runBatch(
			context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
		); !errors.Is(err, ErrContract) {
			t.Fatalf("malformed EXPLAIN error=%v", err)
		}
		assertNoDataCalls(t, driver, statements)
	}
}

func fixtureLowPolicy(t *testing.T) planner.Policy {
	t.Helper()
	policy, err := planner.NewPolicy(planner.Limits{
		RecordStatementCost: 1, RecordRequestCost: 1,
		ExactCountStatementCost: 1, AggregateStatementCost: 1,
		InteractiveTimeout: time.Second, ExactCountTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func assertNoDataCalls(
	t *testing.T,
	driver *fakeDriver,
	statements []statement,
) {
	t.Helper()
	for _, event := range driver.events {
		if strings.HasPrefix(event, "query:") ||
			len(statements) == 2 && event == "row:"+statements[1].sql {
			t.Fatalf("data statement ran after rejection: %v", driver.events)
		}
	}
}

func assertQueryCode(
	t *testing.T,
	err error,
	code restqueryv1.ErrorCode,
) {
	t.Helper()
	var queryError api.QueryError
	if !errors.As(err, &queryError) || queryError.Code() != code {
		t.Fatalf("error=%v code=%s", err, queryError.Code())
	}
}
