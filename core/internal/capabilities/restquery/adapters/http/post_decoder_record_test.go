package httpquery

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestPOSTDecoderDecodesCanonicalRecordQuery(t *testing.T) {
	t.Parallel()
	document := `{"select":["id","status"],` +
		`"where":{"and":[{"field":"status","op":"eq","value":"OPEN"}]},` +
		`"sort":[{"field":"created_at","direction":"desc"}],` +
		`"page":{"size":25,"after":"cursor"},"search":{"term":"acme"},` +
		`"count":"exact","include":[]}`
	request := decodeSuccess(t, document)
	if request.Mode() != api.QueryModeRecord || request.Page().Size() != 25 ||
		request.Count() != api.CountModeExact {
		t.Fatalf("record scalars were not decoded: %#v", request)
	}
	if after, present := request.Page().After(); !present || after != "cursor" {
		t.Fatalf("after = %q, present = %v", after, present)
	}
	if got := request.Select(); len(got) != 2 || got[0] != "id" || got[1] != "status" {
		t.Fatalf("select = %#v", got)
	}
	where, present := request.Where()
	if !present || where.Kind() != api.RequestFilterAnd || len(where.Children()) != 1 {
		t.Fatalf("where = %#v, present = %v", where, present)
	}
	search, present := request.Search()
	if !present || search.Term() != "acme" {
		t.Fatalf("search = %#v, present = %v", search, present)
	}
}
