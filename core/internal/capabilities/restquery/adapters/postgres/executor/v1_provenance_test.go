package executor

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/google/uuid"
)

func TestV1CompilerRejectsPlanAndOrdinalTransplants(t *testing.T) {
	fixture := newV1RunFixture(t)
	tuple, err := fixture.predecessor.MintTuple([]cursor.LastValue{
		v1UUIDLast(t, "40000000-0000-4000-8000-000000000002"),
	})
	if err != nil {
		t.Fatal(err)
	}
	other := newBatchFixture(
		t, api.CountModeExact,
		uuid.MustParse("30000000-0000-4000-8000-000000000099"), fixtureActor,
	)
	if !errors.Is(
		fixture.predecessor.ValidatePlan(other.plan, fixture.ordinal),
		compiler.ErrContract,
	) {
		t.Fatal("same-shape predecessor crossed authorized tenant plan")
	}
	if _, err := fixture.compiler.CompileV1Page(
		other.plan, fixture.ordinal, fixture.predecessor, tuple,
	); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("same-shape tuple crossed authorized tenant plan")
	}
	wrongPlan := newBatchFixtureWithPageSize(
		t, api.CountModeExact, fixtureTenant, fixtureActor, 1,
	).plan
	wrongPagePlan := executorV1PagePlan(t, wrongPlan, 3, 1)
	wrongOrdinal, present, err := wrongPagePlan.PredecessorOrdinal()
	wrongIndex, _ := wrongOrdinal.Index()
	rightIndex, _ := fixture.ordinal.Index()
	if err != nil || !present || wrongIndex != rightIndex {
		t.Fatal("same-index ordinal fixture is invalid")
	}
	if _, err := fixture.compiler.CompileV1Predecessor(
		fixture.plan, wrongOrdinal,
	); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("same-index ordinal with foreign page size was accepted")
	}
}
