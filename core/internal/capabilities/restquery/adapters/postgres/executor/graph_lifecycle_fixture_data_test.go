package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func (fixture *graphFakeFixture) prepareCanonicalData(t *testing.T) {
	t.Helper()
	ids := fixture.ids
	rootStatements, err := statementsFor(fixture.runtime.batch)
	if err != nil {
		t.Fatal(err)
	}
	rootColumns, _ := fixture.runtime.batch.record.Columns()
	root := fixture.drivers[0]
	root.rows[rootStatements[0].sql] = &fakeRows{values: [][]any{
		graphMaterialRow(t, rootColumns, "root-one", ids.rootOne),
		graphMaterialRow(t, rootColumns, "root-two", ids.rootTwo),
	}}
	fixture.explain(root, rootStatements...)
	if fixture.runtime.batch.hasExactCount {
		root.rowValues[rootStatements[1].sql] = int64(7)
	}
	sources := map[string][]uuid.UUID{
		"children": {ids.rootOne, ids.rootTwo},
		"owner":    {ids.rootOne, ids.rootTwo},
		"items":    {ids.alpha, ids.middle},
	}
	edges := canonicalGraphEdges(ids)
	candidates := canonicalGraphCandidates(ids)
	for index := range fixture.runtime.nodes {
		relation, _ := fixture.runtime.nodes[index].compiled.RelationName()
		driver := fixture.edgeDriver(index)
		fixture.prepareEdge(t, driver, index, sources[relation], edges[relation])
		fixture.prepareTarget(
			t, fixture.drivers[index+1], index, candidates[relation],
			canonicalTargetRows(relation, ids),
		)
	}
}

func (fixture *graphFakeFixture) prepareEdge(
	t *testing.T,
	driver *fakeDriver,
	nodeIndex int,
	sources []uuid.UUID,
	rows [][]any,
) {
	t.Helper()
	node := fixture.runtime.nodes[nodeIndex].compiled
	compiled, err := fixture.runtime.compiler.CompileRelationEdge(
		fixture.runtime.plan, fixture.runtime.graph, node, sources,
	)
	if err != nil {
		t.Fatal(err)
	}
	value, err := consumeRelationEdge(
		compiled, fixture.runtime.plan, fixture.runtime.graph, node,
	)
	if err != nil {
		t.Fatal(err)
	}
	relation, _ := node.RelationName()
	fixture.edges[relation] = value
	driver.rows[value.sql] = &fakeRows{values: rows}
	fixture.explain(driver, value)
}

func (fixture *graphFakeFixture) prepareTarget(
	t *testing.T,
	driver *fakeDriver,
	nodeIndex int,
	candidates []uuid.UUID,
	rows []graphTargetRow,
) {
	t.Helper()
	node := fixture.runtime.nodes[nodeIndex].compiled
	compiled, err := fixture.runtime.compiler.CompileIncludeTarget(
		fixture.runtime.plan, fixture.runtime.graph, node, candidates,
	)
	if err != nil {
		t.Fatal(err)
	}
	value, sealed, err := consumeIncludeTarget(
		compiled, fixture.runtime.plan, fixture.runtime.graph, node,
	)
	if err != nil {
		t.Fatal(err)
	}
	columns, _ := sealed.Columns()
	values := make([][]any, len(rows))
	for index, row := range rows {
		values[index] = graphMaterialRow(t, columns, row.label, row.id)
	}
	relation, _ := node.RelationName()
	fixture.targets[relation] = value
	driver.rows[value.sql] = &fakeRows{values: values}
	fixture.explain(driver, value)
}

func (fixture *graphFakeFixture) explain(driver *fakeDriver, values ...statement) {
	for _, value := range values {
		driver.rowValues[value.explainSQL] = explainDocument("1")
	}
}

func (fixture *graphFakeFixture) edgeDriver(nodeIndex int) *fakeDriver {
	parent := fixture.runtime.nodes[nodeIndex].material.parent
	if parent < 0 {
		return fixture.drivers[0]
	}
	return fixture.drivers[parent+1]
}

type graphTargetRow struct {
	label string
	id    uuid.UUID
}

func canonicalTargetRows(relation string, ids graphLifecycleIDs) []graphTargetRow {
	switch relation {
	case "children":
		return []graphTargetRow{{"alpha", ids.alpha}, {"middle", ids.middle}, {"zulu", ids.zulu}}
	case "owner":
		return []graphTargetRow{{"owner-one", ids.owner}}
	case "items":
		return []graphTargetRow{{"item-alpha", ids.itemAlpha}, {"item-middle", ids.itemMiddle}, {"item-zulu", ids.itemZulu}}
	default:
		return nil
	}
}

func graphMaterialRow(
	t *testing.T,
	columns []compiler.ResultColumn,
	label string,
	id uuid.UUID,
) []any {
	t.Helper()
	values := make([]any, len(columns))
	for index, column := range columns {
		kind, err := column.LogicalType()
		if err != nil {
			t.Fatal(err)
		}
		switch kind {
		case api.LogicalTypeText:
			values[index] = label
		case api.LogicalTypeUUID:
			values[index] = [16]byte(id)
		default:
			t.Fatalf("unexpected graph column type %s", kind)
		}
	}
	return values
}
