package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestAggregateStatementAndBatchAreSealedAndRedacted(t *testing.T) {
	measure := fixtureAggregateMeasure(t, "rows", api.AggregateFunctionCount, nil)
	query, err := resolved.NewAggregateQuery(
		[]resolved.AggregateGroup{}, []resolved.AggregateMeasure{measure}, nil,
		[]resolved.SearchTerm{}, nil, []resolved.AggregateSort{}, 10,
	)
	if err != nil {
		t.Fatal(err)
	}
	measures, _ := extractAggregateMeasures(query)
	input := fixtureAggregateInput(
		t, query, []aggregateGroup{}, measures, fixtureAllVisibility(t),
	)
	statement, err := compileAggregate(input)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := newAggregateBatch(statement, input.binding)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{statement, batch} {
		if fmt.Sprint(value) != compilerRedaction {
			t.Fatalf("%T leaked formatting", value)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T marshaled directly", value)
		}
	}
	tamperedStatement := statement
	tamperedStatement.sql += " "
	if !errors.Is(tamperedStatement.Validate(), ErrContract) {
		t.Fatal("tampered aggregate statement validated")
	}
	tamperedBatch := batch
	tamperedBatch.stateSeal[0] ^= 1
	if !errors.Is(tamperedBatch.Validate(), ErrContract) {
		t.Fatal("tampered aggregate batch validated")
	}
}
