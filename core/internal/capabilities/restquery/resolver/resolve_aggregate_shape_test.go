package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func assertAggregateHavingAndSort(
	t *testing.T, aggregate resolved.AggregateQuery,
) {
	t.Helper()
	having, found, err := aggregate.Having()
	if err != nil || !found {
		t.Fatalf("HAVING missing: %v", err)
	}
	predicate, found, err := having.Predicate()
	if err != nil || !found {
		t.Fatalf("HAVING predicate missing: %v", err)
	}
	reference, _ := predicate.Reference()
	kind, _ := reference.Kind()
	ordinal, _ := reference.Ordinal()
	path, _ := reference.Path()
	pathValue, _ := path.Value()
	if kind != api.AggregateReferenceMeasure || ordinal != 0 ||
		pathValue != "/aggregate/having" {
		t.Fatal("HAVING did not resolve to a typed ordinal and canonical path")
	}
	sorts, err := aggregate.Sorts()
	if err != nil || len(sorts) != 2 {
		t.Fatalf("aggregate sorts = %d, err = %v", len(sorts), err)
	}
	firstReference, _ := sorts[0].Reference()
	firstKind, _ := firstReference.Kind()
	firstOrdinal, _ := firstReference.Ordinal()
	firstImplicit, _ := sorts[0].Implicit()
	secondReference, _ := sorts[1].Reference()
	secondKind, _ := secondReference.Kind()
	secondOrdinal, _ := secondReference.Ordinal()
	secondDirection, _ := sorts[1].Direction()
	secondImplicit, _ := sorts[1].Implicit()
	if firstKind != api.AggregateReferenceMeasure || firstOrdinal != 0 || firstImplicit ||
		secondKind != api.AggregateReferenceGroup || secondOrdinal != 0 ||
		secondDirection != api.DirectionAscending || !secondImplicit {
		t.Fatal("aggregate explicit sort or implicit group-tuple tie was lost")
	}
}
