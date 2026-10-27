package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type v1RunFixture struct {
	batch       consumedBatch
	plan        resolved.AuthorizedPlan
	codec       cursor.Codec
	compiler    compiler.Compiler
	ordinal     v1query.PredecessorOrdinal
	predecessor compiler.PredecessorStatement
	page        compiler.RecordStatement
	seek        statement
	pageSQL     statement
	count       statement
}

func newV1RunFixture(t *testing.T) v1RunFixture {
	t.Helper()
	base := newBatchFixture(t, api.CountModeExact, fixtureTenant, fixtureActor)
	pagePlan := executorV1PagePlan(t, base.plan, 2, 2)
	ordinal, present, err := pagePlan.PredecessorOrdinal()
	if err != nil || !present {
		t.Fatal("V1 predecessor ordinal missing")
	}
	postgresCompiler, _ := compiler.New(base.codec)
	predecessor, err := postgresCompiler.CompileV1Predecessor(base.plan, ordinal)
	if err != nil {
		t.Fatal(err)
	}
	last := v1UUIDLast(t, "40000000-0000-4000-8000-000000000002")
	tuple, err := predecessor.MintTuple([]cursor.LastValue{last})
	if err != nil {
		t.Fatal(err)
	}
	page, err := postgresCompiler.CompileV1Page(
		base.plan, ordinal, predecessor, tuple,
	)
	if err != nil {
		t.Fatal(err)
	}
	batch := consumedFixture(t, base)
	seek, _ := predecessorExecutorStatement(predecessor)
	pageSQL, _ := recordStatement(page)
	count, _ := countStatement(batch.exactCount)
	return v1RunFixture{
		batch: batch, plan: base.plan, codec: base.codec, compiler: postgresCompiler,
		ordinal: ordinal, predecessor: predecessor, page: page,
		seek: seek, pageSQL: pageSQL, count: count,
	}
}

func v1UUIDLast(t *testing.T, value string) cursor.LastValue {
	t.Helper()
	resolvedValue, _ := resolved.NewScalarValue(api.LogicalTypeUUID, value)
	last, err := cursor.NewLastValue(resolvedValue)
	if err != nil {
		t.Fatal(err)
	}
	return last
}

func (fixture v1RunFixture) driver(
	t *testing.T, seekCost string, pageCost string,
) *fakeDriver {
	t.Helper()
	record := fixtureRecordRow(
		t, fixture.batch, "Cem", []byte(`{"level":3}`),
		"40000000-0000-4000-8000-000000000003",
	)
	return &fakeDriver{
		owner: "forms",
		rows: map[string]*fakeRows{
			fixture.seek.sql: {
				values: [][]any{{[16]byte(uuid.MustParse(
					"40000000-0000-4000-8000-000000000002",
				))}},
			},
			fixture.pageSQL.sql: {values: [][]any{record}},
		},
		rowValues: map[string]any{
			fixture.seek.explainSQL:    explainDocument(seekCost),
			fixture.pageSQL.explainSQL: explainDocument(pageCost),
			fixture.count.explainSQL:   explainDocument("1"),
			fixture.count.sql:          int64(3),
		},
		rowErrors: map[string]error{}, queryErrors: map[string]error{},
	}
}
