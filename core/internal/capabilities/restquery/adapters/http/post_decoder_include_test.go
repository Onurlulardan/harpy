package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderDecodesRecursiveIncludes(t *testing.T) {
	t.Parallel()
	document := `{"include":[{"relation":"customer","select":["id","name"],` +
		`"where":null,"sort":[{"field":"name","direction":"asc"}],"limit":20,` +
		`"include":[{"relation":"address","select":[],"include":[]}]}]}`
	request := decodeSuccess(t, document)
	includes := request.Includes()
	if len(includes) != 1 || includes[0].Relation() != "customer" {
		t.Fatalf("includes = %#v", includes)
	}
	if limit, present := includes[0].Limit(); !present || limit != 20 {
		t.Fatalf("limit = %d, present = %v", limit, present)
	}
	children := includes[0].Includes()
	if len(children) != 1 || children[0].Relation() != "address" {
		t.Fatalf("children = %#v", children)
	}
	if _, present := children[0].Limit(); present {
		t.Fatal("absent child limit became present")
	}
}

func TestPOSTDecoderEnforcesIncludeDepth(t *testing.T) {
	t.Parallel()
	document := []byte(`{"include":[{"relation":"a","include":[` +
		`{"relation":"b","include":[{"relation":"c"}]}]}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/include/0/include/0/include/0", restqueryv1.IssueIncludeDepthExceeded}})
}

func TestPOSTDecoderEnforcesIncludeEdgeCount(t *testing.T) {
	t.Parallel()
	document := []byte(`{"include":[{"relation":"a"},{"relation":"b"},` +
		`{"relation":"c"},{"relation":"d"}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/include/3", restqueryv1.IssueIncludeEdgeCountExceeded}})
}

func TestPOSTDecoderUsesIncludeLimitIssueCode(t *testing.T) {
	t.Parallel()
	document := []byte(`{"include":[{"relation":"items","limit":101}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/include/0/limit", restqueryv1.IssueIncludeResultCountExceeded}})
}

func TestPOSTDecoderRejectsDuplicateSiblingRelation(t *testing.T) {
	t.Parallel()
	document := []byte(`{"include":[{"relation":"owner"},{"relation":"owner"}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/include/1/relation", restqueryv1.IssueNameDuplicate}})
}
