package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderMeasuresSearchAfterNFCNormalization(t *testing.T) {
	t.Parallel()
	decomposed := strings.Repeat("e\u0301", 128)
	request := decodeSuccess(t, `{"search":{"term":"`+decomposed+`"}}`)
	search, present := request.Search()
	if !present || search.Term() != strings.Repeat("é", 128) {
		t.Fatalf("normalized search = %#v, present = %v", search, present)
	}
}

func TestPOSTDecoderRejectsNormalizedSearchAboveLimit(t *testing.T) {
	t.Parallel()
	decomposed := strings.Repeat("e\u0301", 129)
	document := []byte(`{"search":{"term":"` + decomposed + `"}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/search/term", restqueryv1.IssueSearchSizeExceeded}})
}
