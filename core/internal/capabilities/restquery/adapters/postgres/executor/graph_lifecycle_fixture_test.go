package executor

import (
	"sync"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type graphCallLog struct {
	mutex sync.Mutex
	calls []string
}

func (log *graphCallLog) append(kind, sql string) {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	log.calls = append(log.calls, kind+"|"+sql)
}

func (log *graphCallLog) snapshot() []string {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	return append([]string(nil), log.calls...)
}

type graphFakeFixture struct {
	runtime *graphRuntime
	drivers []*fakeDriver
	ids     graphLifecycleIDs
	edges   map[string]statement
	targets map[string]statement
	log     *graphCallLog
}

func newGraphFakeFixture(t *testing.T, count api.CountMode) *graphFakeFixture {
	t.Helper()
	fixture := newGraphFakeRuntimeFixture(t, count, 2)
	fixture.prepareCanonicalData(t)
	return fixture
}

func newGraphFakeRuntimeFixture(
	t *testing.T,
	count api.CountMode,
	pageSize int,
) *graphFakeFixture {
	t.Helper()
	planFixture := newGraphPlanFixture(t, count, pageSize)
	codec := fixtureCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := postgresCompiler.CompileRecordBatch(planFixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := consumeBatch(batch, planFixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := postgresCompiler.CompileIncludeGraph(planFixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	stableAt, err := rootStableColumn(planFixture.plan, consumed.record)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := newGraphRuntime(
		planFixture.plan, graph, postgresCompiler, consumed, codec,
		planner.HardPolicy(), stableAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &graphFakeFixture{
		runtime: runtime, ids: canonicalGraphIDs(), log: &graphCallLog{},
		edges: map[string]statement{}, targets: map[string]statement{},
	}
	fixture.drivers = makeGraphDrivers(t, runtime, fixture.log)
	return fixture
}

func makeGraphDrivers(
	t *testing.T,
	runtime *graphRuntime,
	log *graphCallLog,
) []*fakeDriver {
	t.Helper()
	owners := make([]string, len(runtime.nodes)+1)
	rootCatalog, _ := runtime.batch.record.Catalog()
	owners[0], _ = rootCatalog.Owner()
	for index, node := range runtime.nodes {
		owners[index+1], _ = node.compiled.Owner()
	}
	drivers := make([]*fakeDriver, len(owners))
	for index, owner := range owners {
		drivers[index] = &fakeDriver{
			owner: owner, rows: map[string]*fakeRows{}, rowValues: map[string]any{},
			rowErrors: map[string]error{}, queryErrors: map[string]error{},
			onCall: log.append,
		}
	}
	return drivers
}
