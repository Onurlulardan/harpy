package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestGETDecoderRequiresNFCForStructuralNames(t *testing.T) {
	requireGETError(t, "select[0]=e%CC%81", restqueryv1.ErrorQueryInvalid,
		[]expectedGETIssue{{"/select/0", restqueryv1.IssueNFCRequired}})
	requireGETError(t,
		"filter[0][valu%65%CC%81]=x&filter[0][field]=f&filter[0][op]=eq",
		restqueryv1.ErrorQueryInvalid,
		[]expectedGETIssue{{"/filter/0/valué", restqueryv1.IssueNFCRequired}},
	)
}

func TestGETDecoderPreservesUntypedJSONLookingUnicode(t *testing.T) {
	rawQuery := "filter[0][field]=payload&filter[0][op]=eq&" +
		"filter[0][value]=%7B%22e%22%3A%22e%CC%81%22%7D"
	request := decodeGETSuccess(t, rawQuery)
	where, _ := request.Where()
	predicate, _ := where.Children()[0].Predicate()
	value := getTextValue(t, predicate.Values()[0])
	if value != "{\"e\":\"e\u0301\"}" {
		t.Fatalf("untyped GET literal was normalized: %q", value)
	}
}
