package api

import "testing"

func TestQueryRequestValidateRejectsZeroAndAggregateMode(t *testing.T) {
	if err := (QueryRequest{}).Validate(); err == nil {
		t.Fatal("zero query request validated")
	}
	request := mustDeepRecordQueryRequest(t)
	if err := request.Validate(); err != nil {
		t.Fatalf("constructed request is invalid: %v", err)
	}
	request.mode = QueryModeAggregate
	if err := request.Validate(); err == nil {
		t.Fatal("aggregate mode validated without its constructor")
	}
}

func TestQueryRequestValidateRejectsMalformedRecordMembers(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*QueryRequest)
	}{
		{"non-canonical collections", func(value *QueryRequest) { value.sorts = nil }},
		{"duplicate select", func(value *QueryRequest) { value.selectFields = []string{"id", "id"} }},
		{"filter", func(value *QueryRequest) { value.where = &RequestFilter{} }},
		{"sort", func(value *QueryRequest) { value.sorts[0] = RequestSort{} }},
		{"page", func(value *QueryRequest) { value.page.hasAfter, value.page.after = true, "" }},
		{"search", func(value *QueryRequest) { value.search.term = "bad\x00term" }},
		{"count", func(value *QueryRequest) { value.count = 0 }},
		{"deep include", func(value *QueryRequest) {
			value.includes[0].children[0].relation = ""
		}},
		{"cached projection", func(value *QueryRequest) {
			value.includes[0].children[0].projectedFields++
		}},
		{"total projection", func(value *QueryRequest) {
			value.selectFields = makeRequestNames(50, "root")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := mustDeepRecordQueryRequest(t)
			test.mutate(&request)
			if err := request.Validate(); err == nil {
				t.Fatal("malformed request validated")
			}
		})
	}
}

func mustDeepRecordQueryRequest(t *testing.T) QueryRequest {
	t.Helper()
	leaf := mustRequestInclude(t, RequestIncludeInput{
		Relation: "items", Select: []string{"id"},
	})
	parent := mustRequestInclude(t, RequestIncludeInput{
		Relation: "customer", Select: []string{"id"}, Include: []RequestInclude{leaf},
	})
	where := mustPredicate(t)
	search := mustRequestSearch(t, "term")
	request, err := NewRecordQueryRequest(RecordQueryRequestInput{
		Select: []string{"id"}, Where: &where,
		Sort: []RequestSort{mustRequestSort(t, "id")},
		Page: mustRequestPage(t, 50, nil), Search: &search,
		Count: CountModeNone, Include: []RequestInclude{parent},
	})
	if err != nil {
		t.Fatal(err)
	}
	return request
}
