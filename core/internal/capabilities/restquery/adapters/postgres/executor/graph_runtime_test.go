package executor

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestGraphRuntimeBuildsOneProgramNodePerDependencyNode(t *testing.T) {
	fixture := newGraphPlanFixture(t, api.CountModeExact, 2)
	codec := fixtureCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := postgresCompiler.CompileRecordBatch(fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := consumeBatch(batch, fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	stableIndex, err := rootStableColumn(fixture.plan, consumed.record)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := postgresCompiler.CompileIncludeGraph(fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := newGraphRuntime(
		fixture.plan, graph, postgresCompiler, consumed,
		codec, planner.HardPolicy(), stableIndex,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(runtime.nodes) != 3 ||
		!reflect.DeepEqual(runtime.barrier.expected, []int{1, 1, 2, 1, 1}) {
		t.Fatalf("nodes=%d expected=%v", len(runtime.nodes), runtime.barrier.expected)
	}
}
