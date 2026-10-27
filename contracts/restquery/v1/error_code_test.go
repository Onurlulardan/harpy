package restqueryv1

import (
	"reflect"
	"testing"
)

func TestErrorCatalogIsExhaustive(t *testing.T) {
	t.Parallel()
	expected := []ErrorDefinition{
		{Code: "QUERY_AGGREGATE_OVERFLOW", HTTPStatus: 422},
		{Code: "QUERY_BUDGET_EXCEEDED", HTTPStatus: 422},
		{Code: "QUERY_COUNT_BUDGET_EXCEEDED", HTTPStatus: 422},
		{Code: "QUERY_COUNT_TIMEOUT", HTTPStatus: 504},
		{Code: "QUERY_CURSOR_EXPIRED", HTTPStatus: 400},
		{Code: "QUERY_CURSOR_INVALID", HTTPStatus: 400},
		{Code: "QUERY_CURSOR_STALE", HTTPStatus: 409},
		{Code: "QUERY_DUPLICATE_MEMBER", HTTPStatus: 422},
		{Code: "QUERY_FIELD_FORBIDDEN", HTTPStatus: 403},
		{Code: "QUERY_FORBIDDEN", HTTPStatus: 403},
		{Code: "QUERY_INVALID", HTTPStatus: 422},
		{Code: "QUERY_PARAMETER_INVALID", HTTPStatus: 422},
		{Code: "QUERY_TIMEOUT", HTTPStatus: 504},
		{Code: "QUERY_TRANSPORT_INVALID", HTTPStatus: 400},
		{Code: "RELATION_CARDINALITY_VIOLATION", HTTPStatus: 409},
	}
	if actual := ErrorCatalog(); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("error catalog drifted\nwant: %#v\n got: %#v", expected, actual)
	}
}

func TestErrorCatalogReturnsCopyAndStatus(t *testing.T) {
	t.Parallel()
	catalog := ErrorCatalog()
	catalog[0].HTTPStatus = 200
	status, found := ErrorStatus(ErrorQueryAggregateOverflow)
	if !found || status != 422 {
		t.Fatalf("catalog storage was mutable: status=%d found=%v", status, found)
	}
	if _, found := ErrorStatus("QUERY_NOT_PUBLISHED"); found {
		t.Fatal("unknown error code received a status")
	}
}
