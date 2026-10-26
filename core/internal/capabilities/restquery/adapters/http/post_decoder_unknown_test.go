package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderReportsUnknownMembersAtEverySemanticObject(t *testing.T) {
	t.Parallel()
	document := []byte(`{"a/b~c":1,"page":{"size":50,"after":null,"extra":1},` +
		`"sort":[{"field":"id","direction":"asc","x":1}],` +
		`"include":[{"relation":"owner","unknown":true}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{
			{"/a~1b~0c", restqueryv1.IssueMemberUnknown},
			{"/include/0/unknown", restqueryv1.IssueMemberUnknown},
			{"/page/extra", restqueryv1.IssueMemberUnknown},
			{"/sort/0/x", restqueryv1.IssueMemberUnknown},
		})
}

func TestPOSTDecoderReportsUnknownAggregateMembers(t *testing.T) {
	t.Parallel()
	document := []byte(`{"aggregate":{"mystery":1,"measures":[` +
		`{"name":"rows","function":"count","extra":true}]}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{
			{"/aggregate/measures/0/extra", restqueryv1.IssueMemberUnknown},
			{"/aggregate/mystery", restqueryv1.IssueMemberUnknown},
		})
}
