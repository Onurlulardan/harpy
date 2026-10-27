package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestPOSTDecoderAppliesRecordDefaults(t *testing.T) {
	t.Parallel()
	request := decodeSuccess(t, `{}`)
	if request.Mode() != api.QueryModeRecord {
		t.Fatalf("mode = %v", request.Mode())
	}
	if request.Select() == nil || len(request.Select()) != 0 {
		t.Fatalf("select = %#v", request.Select())
	}
	if request.Sort() == nil || len(request.Sort()) != 0 {
		t.Fatalf("sort = %#v", request.Sort())
	}
	if request.Includes() == nil || len(request.Includes()) != 0 {
		t.Fatalf("include = %#v", request.Includes())
	}
	if request.Page().Size() != 50 {
		t.Fatalf("page size = %d", request.Page().Size())
	}
	if _, present := request.Page().After(); present {
		t.Fatal("default cursor is present")
	}
	if request.Count() != api.CountModeNone {
		t.Fatalf("count = %v", request.Count())
	}
	if _, present := request.Where(); present {
		t.Fatal("where is present")
	}
	if _, present := request.Search(); present {
		t.Fatal("search is present")
	}
	if _, present := request.Aggregate(); present {
		t.Fatal("aggregate is present")
	}
}

func TestPOSTDecoderPreservesNullAndEmptyRecordMembers(t *testing.T) {
	t.Parallel()
	document := `{"select":[],"where":null,"sort":[],"page":{"size":50,"after":null},` +
		`"search":null,"count":"none","include":[],"aggregate":null}`
	request := decodeSuccess(t, document)
	if request.Mode() != api.QueryModeRecord || len(request.Select()) != 0 ||
		len(request.Sort()) != 0 || len(request.Includes()) != 0 {
		t.Fatalf("record defaults were not preserved: %#v", request)
	}
}

func TestPOSTDecoderRejectsEmptyAggregateDeclaration(t *testing.T) {
	t.Parallel()
	requireQueryError(t, []byte(`{"aggregate":{}}`), restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/aggregate", restqueryv1.IssueValueNotAllowed}})
}
