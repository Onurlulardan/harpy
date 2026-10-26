package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestPOSTDecoderDecodesCanonicalAggregateQuery(t *testing.T) {
	t.Parallel()
	document := `{"where":{"field":"active","op":"eq","value":true},` +
		`"search":{"term":"acme"},"aggregate":{"group_by":["status"],` +
		`"measures":[{"name":"rows","function":"count"},` +
		`{"name":"revenue","function":"sum","field":"total"}],` +
		`"having":{"measure":"revenue","op":"gte","value":"1000.00"},` +
		`"sort":[{"measure":"revenue","direction":"desc"}],"limit":25}}`
	request := decodeSuccess(t, document)
	aggregate, present := request.Aggregate()
	if request.Mode() != api.QueryModeAggregate || !present || aggregate.Limit() != 25 {
		t.Fatalf("aggregate = %#v, present = %v", aggregate, present)
	}
	if groups := aggregate.GroupBy(); len(groups) != 1 || groups[0] != "status" {
		t.Fatalf("groups = %#v", groups)
	}
	if measures := aggregate.Measures(); len(measures) != 2 ||
		measures[0].Function() != api.AggregateFunctionCount || measures[1].Name() != "revenue" {
		t.Fatalf("measures = %#v", measures)
	}
	having, present := aggregate.Having()
	if !present || having.PredicateCount() != 1 {
		t.Fatalf("having = %#v, present = %v", having, present)
	}
}

func TestPOSTDecoderForbidsRecordMembersInAggregateMode(t *testing.T) {
	t.Parallel()
	document := []byte(`{"select":[],"sort":[],"page":null,"include":[],` +
		`"count":"none","aggregate":{}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{
			{"/aggregate", restqueryv1.IssueValueNotAllowed},
			{"/count", restqueryv1.IssueMemberForbidden},
			{"/include", restqueryv1.IssueMemberForbidden},
			{"/page", restqueryv1.IssueMemberForbidden},
			{"/select", restqueryv1.IssueMemberForbidden},
			{"/sort", restqueryv1.IssueMemberForbidden},
		})
}

func TestPOSTDecoderEnforcesMeasureFieldUnion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		measure  string
		expected expectedQueryIssue
	}{
		{"count field", `{"name":"rows","function":"count","field":"id"}`,
			expectedQueryIssue{"/aggregate/measures/0/field", restqueryv1.IssueMemberForbidden}},
		{"sum missing field", `{"name":"sum_total","function":"sum"}`,
			expectedQueryIssue{"/aggregate/measures/0/field", restqueryv1.IssueMemberRequired}},
		{"unknown function", `{"name":"rows","function":"median","field":"total"}`,
			expectedQueryIssue{"/aggregate/measures/0/function", restqueryv1.IssueFunctionUnknown}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document := []byte(`{"aggregate":{"measures":[` + test.measure + `]}}`)
			requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
				[]expectedQueryIssue{test.expected})
		})
	}
}

func TestPOSTDecoderEnforcesAggregateSortReferenceUnion(t *testing.T) {
	t.Parallel()
	document := []byte(`{"aggregate":{"group_by":["status"],` +
		`"measures":[{"name":"rows","function":"count"}],` +
		`"sort":[{"group":"status","measure":"rows","direction":"asc"}]}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/aggregate/sort/0/measure", restqueryv1.IssueMemberForbidden}})
}

func TestPOSTDecoderRejectsUndeclaredHavingReference(t *testing.T) {
	t.Parallel()
	document := []byte(`{"aggregate":{"measures":[{"name":"rows","function":"count"}],` +
		`"having":{"measure":"missing","op":"eq","value":"1"}}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/aggregate/having/measure", restqueryv1.IssueReferenceInvalid}})
}

func TestPOSTDecoderEnforcesHavingReferenceUnion(t *testing.T) {
	t.Parallel()
	document := []byte(`{"aggregate":{"group_by":["status"],` +
		`"measures":[{"name":"rows","function":"count"}],` +
		`"having":{"group":"status","measure":"rows","op":"eq","value":"OPEN"}}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/aggregate/having/measure", restqueryv1.IssueMemberForbidden}})
}
