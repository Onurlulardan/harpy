package v1query

import (
	"errors"
	"slices"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type staticManifestSource struct {
	snapshot ManifestSnapshot
	err      error
}

func (source staticManifestSource) VerifiedV1QueryManifest() (ManifestSnapshot, error) {
	return source.snapshot, source.err
}

func testManifest() ManifestSnapshot {
	return ManifestSnapshot{
		ManifestID:  "51000000-0000-4000-8000-000000000001",
		OwnerModule: "forms", Domain: "forms",
		DefaultLimit: 10, MaximumPageSize: 500,
		DefaultSortField: "created_at", DefaultSortDirection: "desc",
		FilterFields: map[string]FilterFieldBinding{
			"status":   {Name: "status", Shape: FilterShapeScalar},
			"priority": {Name: "priority", Shape: FilterShapeScalar},
			"tags":     {Name: "tags", Shape: FilterShapeCollection},
			"legacy_status": {
				Name: "status", Shape: FilterShapeScalar, AliasID: "field.status.v1",
			},
		},
		SortFields: map[string]SortFieldBinding{
			"created_at": {Name: "created_at"},
			"status":     {Name: "status"},
			"legacy_status": {
				Name: "status", AliasID: "sort.status.v1",
			},
		},
		OperatorAliases: map[string]OperatorBinding{
			"same_as": {Operator: "eq", AliasID: "operator.same-as.v1"},
		},
		QueryKeyAliases: map[string]QueryKeyBinding{
			"q":        {Key: "search", AliasID: "query.search.v1"},
			"per_page": {Key: "limit", AliasID: "query.limit.v1"},
		},
	}
}

func mustDecoder(t *testing.T) decoder {
	t.Helper()
	decoder, err := newDecoder(staticManifestSource{snapshot: testManifest()})
	if err != nil {
		t.Fatalf("newDecoder() error = %v", err)
	}
	return decoder
}

func requireQueryError(
	t *testing.T, err error, code restqueryv1.ErrorCode,
	path string, issueCode restqueryv1.IssueCode,
) {
	t.Helper()
	var queryError api.QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("error = %T %v, want api.QueryError", err, err)
	}
	if queryError.Code() != code {
		t.Fatalf("error code = %s, want %s", queryError.Code(), code)
	}
	issues := queryError.Issues()
	for _, issue := range issues {
		if issue.Path == path && issue.Code == string(issueCode) {
			return
		}
	}
	t.Fatalf("issues = %#v, want {%q,%q}", issues, path, issueCode)
}

func requirePredicate(
	t *testing.T, filter api.RequestFilter, field string,
	operator api.Operator, values []string,
) {
	t.Helper()
	predicate, found := filter.Predicate()
	if !found || predicate.Field() != field || predicate.Operator() != operator {
		t.Fatalf("predicate = %#v, want %s/%s", predicate, field, operator)
	}
	actual := make([]string, 0, len(predicate.Values()))
	for _, literal := range predicate.Values() {
		value, text, err := literal.TextValue()
		if err != nil || !text {
			t.Fatalf("literal is not valid text: %v", err)
		}
		actual = append(actual, value)
	}
	if !slices.Equal(actual, values) {
		t.Fatalf("values = %#v, want %#v", actual, values)
	}
}
