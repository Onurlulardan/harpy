package systemresourcequery

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestRelationLookupRequestUsesBoundedCanonicalIDPredicate(t *testing.T) {
	ids := []uuid.UUID{
		uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000098"),
		uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000099"),
	}
	opaque, err := NewRelationLookupRequest(ids, []string{"id", "name"})
	if err != nil {
		t.Fatal(err)
	}
	request, ok := opaque.(api.QueryRequest)
	where, present := request.Where()
	predicate, predicatePresent := where.Predicate()
	if !ok || !present || !predicatePresent ||
		request.Page().Size() != len(ids) ||
		predicate.Field() != "id" || predicate.Operator() != api.OperatorIn ||
		len(predicate.Values()) != len(ids) {
		t.Fatalf("relation query drifted: ok=%v present=%v predicate=%v page=%d",
			ok, present, predicatePresent, request.Page().Size())
	}
}

func TestRelationLookupRequestEnforcesBatchBound(t *testing.T) {
	ids := make([]uuid.UUID, RelationLookupBatchSize+1)
	for index := range ids {
		ids[index] = uuid.New()
	}
	if _, err := NewRelationLookupRequest(ids, []string{"id"}); err == nil {
		t.Fatal("oversized relation query was accepted")
	}
}
