//go:build integration

package executor

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (gate *pg18Gate) verifyV1Pages(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	wants := map[int64][]string{
		2: {
			"40000000-0000-4000-8000-000000000003",
			"40000000-0000-4000-8000-000000000004",
		},
		3: {
			"40000000-0000-4000-8000-000000000005",
			"40000000-0000-4000-8000-000000000006",
		},
	}
	for page, want := range wants {
		accepted := pg18V1Plan(t,
			"page="+strconv.FormatInt(page, 10)+"&limit=2", pg18TenantA)
		result := gate.pg18ExecuteV1(t, executor, pg18TenantA, accepted)
		actual := pg18ResultIDs(t, result)
		oracle := gate.v1OffsetOracle(t, int((page-1)*2))
		if !slices.Equal(actual, want) || !slices.Equal(actual, oracle) {
			t.Fatalf("page=%d actual=%v want=%v oracle=%v", page, actual, want, oracle)
		}
		pg18AssertExactTotal(t, result, 6)
	}
	seekQueries := 0
	for _, call := range port.observation.snapshot() {
		if call.kind == "query" && strings.Contains(call.statement, " OFFSET ") {
			seekQueries++
		}
	}
	if seekQueries != 2 || port.observation.opened.Load() != port.observation.closed.Load() {
		t.Fatalf("seek queries=%d rows=%d/%d", seekQueries,
			port.observation.opened.Load(), port.observation.closed.Load())
	}
}

func (gate *pg18Gate) verifyV1Missing(t *testing.T) {
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	accepted := pg18V1Plan(t, "page=5&limit=2", pg18TenantA)
	result := gate.pg18ExecuteV1(t, executor, pg18TenantA, accepted)
	if identifiers := pg18ResultIDs(t, result); len(identifiers) != 0 {
		t.Fatalf("missing predecessor fell back: %v", identifiers)
	}
	pg18AssertExactTotal(t, result, 6)
	dataQueries := 0
	for _, call := range port.observation.snapshot() {
		if call.kind == "query" {
			dataQueries++
		}
	}
	if dataQueries != 1 {
		t.Fatalf("missing predecessor executed %d data row queries", dataQueries)
	}
}
