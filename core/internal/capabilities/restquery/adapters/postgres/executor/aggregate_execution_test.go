package executor

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/jackc/pgx/v5/pgtype"
	"math/big"
)

func TestRunAggregateAssemblesDeclarationOrderAndNulls(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	driver, candidate := readyAggregateDriver(t, batch, [][]any{
		{"OPEN", int64(2), pgtype.Numeric{Int: big.NewInt(12345), Exp: -2, Valid: true}},
		{nil, int64(1), nil},
	}, "1")
	result, err := runAggregate(
		context.Background(), driver, batch, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := result.Groups()
	if err != nil || len(groups) != 2 {
		t.Fatalf("groups=%q err=%v", groups, err)
	}
	want := []string{
		`{"key":{"status":"OPEN"},"measures":{"row_count":"2","revenue":"123.45"}}`,
		`{"key":{"status":null},"measures":{"row_count":"1","revenue":null}}`,
	}
	for index := range want {
		if string(groups[index]) != want[index] {
			t.Fatalf("group %d=%s", index, groups[index])
		}
	}
	if !driver.rows[candidate.sql].closed {
		t.Fatal("aggregate rows were not closed")
	}
	explainIndex := eventIndex(driver.events, "row:"+candidate.explainSQL)
	dataIndex := eventIndex(driver.events, "query:"+candidate.sql)
	if explainIndex < 0 || dataIndex <= explainIndex {
		t.Fatalf("execution order=%v", driver.events)
	}
}

func TestRunAggregateEmptyGlobalAndGroupedSemantics(t *testing.T) {
	globalFixture := newAggregateFixture(t, false)
	global := consumedAggregateFixture(t, globalFixture)
	driver, _ := readyAggregateDriver(t, global, [][]any{{int64(0), nil}}, "1")
	result, err := runAggregate(context.Background(), driver, global, planner.HardPolicy())
	if err != nil {
		t.Fatal(err)
	}
	groups, _ := result.Groups()
	want := `{"key":{},"measures":{"row_count":"0","revenue":null}}`
	if len(groups) != 1 || string(groups[0]) != want {
		t.Fatalf("global groups=%q", groups)
	}
	groupedFixture := newAggregateFixture(t, true)
	grouped := consumedAggregateFixture(t, groupedFixture)
	driver, _ = readyAggregateDriver(t, grouped, [][]any{}, "1")
	result, err = runAggregate(context.Background(), driver, grouped, planner.HardPolicy())
	if err != nil {
		t.Fatal(err)
	}
	groups, _ = result.Groups()
	if len(groups) != 0 {
		t.Fatalf("grouped empty=%q", groups)
	}
}
