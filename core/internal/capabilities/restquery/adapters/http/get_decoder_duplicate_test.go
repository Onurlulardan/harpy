package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestGETDecoderDetectsDuplicateAfterPercentDecode(t *testing.T) {
	requireGETError(t,
		"select[0]=a&select%5B0%5D=b",
		restqueryv1.ErrorQueryDuplicateMember,
		[]expectedGETIssue{{"/select/0", restqueryv1.IssueMemberDuplicate}},
	)
}

func TestGETDecoderDetectsDuplicateAfterNameNFC(t *testing.T) {
	requireGETError(t,
		"e%CC%81=x&%C3%A9=y",
		restqueryv1.ErrorQueryDuplicateMember,
		[]expectedGETIssue{{"/é", restqueryv1.IssueMemberDuplicate}},
	)
}

func TestGETDecoderSortsDuplicatePointers(t *testing.T) {
	rawQuery := "sort[0][field]=a&select[0]=x&" +
		"sort%5B0%5D%5Bfield%5D=b&select%5B0%5D=y"
	requireGETError(t, rawQuery, restqueryv1.ErrorQueryDuplicateMember,
		[]expectedGETIssue{
			{"/select/0", restqueryv1.IssueMemberDuplicate},
			{"/sort/0/field", restqueryv1.IssueMemberDuplicate},
		})
}
