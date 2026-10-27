//go:build integration

package executor

import (
	"bytes"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func (gate *pg18Gate) verifyAggregateExecution(t *testing.T) {
	gate.prepareAggregateFixture(t)
	port := pg18NewOwnerPort(t, "forms")
	executor := pg18NewExecutor(t, gate.codec, planner.HardPolicy(), port)
	t.Run("half even once", func(t *testing.T) {
		gate.verifyAggregateRounding(t, executor, port.observation)
	})
	t.Run("hidden overflow is atomic", func(t *testing.T) {
		gate.verifyAggregateOverflow(t, executor, port.observation)
	})
	t.Run("global and grouped empty", func(t *testing.T) {
		gate.verifyAggregateEmpty(t, executor, port.observation)
	})
}

func (gate *pg18Gate) verifyAggregateRounding(
	t *testing.T, executor Executor, observation *pg18Observation,
) {
	amount := "amount"
	average, _ := api.NewRequestMeasure("average", api.AggregateFunctionAvg, &amount)
	plan := pg18AggregatePlan(t, pg18AggregatePlanInput{
		groups: []string{"status"}, measures: []api.RequestMeasure{average}, limit: 20,
	})
	result, err := gate.runAggregate(t, executor, plan)
	if err != nil {
		t.Fatalf("%v; database=%v", err, observation.errorSnapshot())
	}
	found := map[string]string{}
	for _, document := range pg18AggregateDocuments(t, result) {
		status, present := pg18JSONString(document.Key["status"])
		average, finite := pg18JSONString(document.Measures["average"])
		if present && finite {
			found[status] = average
		}
	}
	want := map[string]string{
		"P_EVEN": "0.000000", "P_ODD": "0.000002",
		"N_EVEN": "0.000000", "N_ODD": "-0.000002",
	}
	if len(found) != len(want) {
		t.Fatalf("rounded groups = %#v", found)
	}
	for status, expected := range want {
		if found[status] != expected {
			t.Fatalf("%s average = %q, want %q", status, found[status], expected)
		}
	}
}

func (gate *pg18Gate) verifyAggregateOverflow(
	t *testing.T, executor Executor, observation *pg18Observation,
) {
	quantity := "quantity"
	sum, _ := api.NewRequestMeasure("quantity_sum", api.AggregateFunctionSum, &quantity)
	literal, _ := api.NewTextLiteral("SAFE")
	having, _ := api.NewRequestHavingPredicate(
		api.AggregateReferenceGroup, "status", api.OperatorEq,
		[]api.Literal{literal},
	)
	sort, _ := api.NewRequestAggregateSort(
		api.AggregateReferenceGroup, "status", api.DirectionAscending,
	)
	plan := pg18AggregatePlan(t, pg18AggregatePlanInput{
		groups: []string{"status"}, measures: []api.RequestMeasure{sum},
		having: &having, sorts: []api.RequestAggregateSort{sort}, limit: 1,
	})
	_, err := gate.runAggregate(t, executor, plan)
	if err != nil {
		t.Logf("aggregate overflow database errors: %v", observation.errorSnapshot())
	}
	pg18AssertQueryCode(t, err, restqueryv1.ErrorQueryAggregateOverflow)
	count, _ := api.NewRequestMeasure("row_count", api.AggregateFunctionCount, nil)
	recovery := pg18AggregatePlan(t, pg18AggregatePlanInput{
		measures: []api.RequestMeasure{count}, where: pg18StatusFilter(t, "SAFE"), limit: 1,
	})
	result, err := gate.runAggregate(t, executor, recovery)
	if err != nil {
		t.Fatalf("post-overflow transaction failed: %v", err)
	}
	documents := pg18AggregateDocuments(t, result)
	if len(documents) != 1 || string(documents[0].Measures["row_count"]) != `"1"` {
		t.Fatalf("post-overflow result = %#v", documents)
	}
}

func (gate *pg18Gate) verifyAggregateEmpty(
	t *testing.T, executor Executor, observation *pg18Observation,
) {
	quantity := "quantity"
	count, _ := api.NewRequestMeasure("row_count", api.AggregateFunctionCount, nil)
	sum, _ := api.NewRequestMeasure("quantity_sum", api.AggregateFunctionSum, &quantity)
	where := pg18StatusFilter(t, "MISSING")
	global := pg18AggregatePlan(t, pg18AggregatePlanInput{
		measures: []api.RequestMeasure{count, sum}, where: where, limit: 5,
	})
	result, err := gate.runAggregate(t, executor, global)
	if err != nil {
		t.Fatalf("%v; database=%v", err, observation.errorSnapshot())
	}
	documents := pg18AggregateDocuments(t, result)
	if len(documents) != 1 || len(documents[0].Key) != 0 ||
		string(documents[0].Measures["row_count"]) != `"0"` ||
		!bytes.Equal(documents[0].Measures["quantity_sum"], []byte("null")) {
		t.Fatalf("global empty = %#v", documents)
	}
	grouped := pg18AggregatePlan(t, pg18AggregatePlanInput{
		groups: []string{"status"}, measures: []api.RequestMeasure{count},
		where: where, limit: 5,
	})
	result, err = gate.runAggregate(t, executor, grouped)
	if err != nil {
		t.Fatalf("%v; database=%v", err, observation.errorSnapshot())
	}
	if groups, err := result.Groups(); err != nil || len(groups) != 0 {
		t.Fatalf("grouped empty = %q, %v", groups, err)
	}
}
