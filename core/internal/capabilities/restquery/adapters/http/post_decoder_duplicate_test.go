package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderRejectsAllDuplicateMembersBeforeMapDecode(t *testing.T) {
	t.Parallel()
	document := []byte(`{"z":1,"z":2,"where":{"field":"payload","op":"eq",` +
		`"value":{"secret":1,"secret":2}},"sort":[{"field":"id","field":"name"}]}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryDuplicateMember,
		[]expectedQueryIssue{
			{"/sort/0/field", restqueryv1.IssueMemberDuplicate},
			{"/where/value/secret", restqueryv1.IssueMemberDuplicate},
			{"/z", restqueryv1.IssueMemberDuplicate},
		})
}

func TestPOSTDecoderEscapesDuplicateMemberPointer(t *testing.T) {
	t.Parallel()
	requireQueryError(t, []byte(`{"a/b~c":1,"a/b~c":2}`),
		restqueryv1.ErrorQueryDuplicateMember,
		[]expectedQueryIssue{{"/a~1b~0c", restqueryv1.IssueMemberDuplicate}})
}

func TestPOSTDecoderAllowsArbitraryLiteralObjectMembers(t *testing.T) {
	t.Parallel()
	request := decodeSuccess(t, `{"where":{"field":"payload","op":"eq",`+
		`"value":{"arbitrary":{"nested":true}}}}`)
	where, present := request.Where()
	if !present {
		t.Fatal("where is absent")
	}
	predicate, present := where.Predicate()
	if !present {
		t.Fatalf("literal was not preserved: %#v", predicate.Values())
	}
	value, jsonPresent, err := predicate.Values()[0].JSONValue()
	if err != nil || !jsonPresent || string(value) != `{"arbitrary":{"nested":true}}` {
		t.Fatalf("literal was not preserved: %#v", predicate.Values())
	}
}
