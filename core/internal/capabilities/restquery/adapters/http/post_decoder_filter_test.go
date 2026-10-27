package httpquery

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderRejectsUnknownOperator(t *testing.T) {
	t.Parallel()
	document := []byte(`{"where":{"field":"status","op":"EQ","value":"OPEN"}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/where/op", restqueryv1.IssueOperatorUnknown}})
}

func TestPOSTDecoderEnforcesPredicateArity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		document string
		issues   []expectedQueryIssue
	}{
		{
			"one uses values",
			`{"where":{"field":"status","op":"eq","values":["OPEN"]}}`,
			[]expectedQueryIssue{
				{"/where/value", restqueryv1.IssueMemberRequired},
				{"/where/values", restqueryv1.IssueArityMismatch},
			},
		},
		{
			"zero carries value",
			`{"where":{"field":"deleted_at","op":"is_null","value":"x"}}`,
			[]expectedQueryIssue{{"/where/value", restqueryv1.IssueArityMismatch}},
		},
		{
			"pair has one value",
			`{"where":{"field":"total","op":"between","values":["1"]}}`,
			[]expectedQueryIssue{{"/where/values", restqueryv1.IssueArityMismatch}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			requireQueryError(t, []byte(test.document), restqueryv1.ErrorQueryInvalid, test.issues)
		})
	}
}

func TestPOSTDecoderRejectsNullPredicateLiteral(t *testing.T) {
	t.Parallel()
	document := []byte(`{"where":{"field":"status","op":"eq","value":null}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/where/value", restqueryv1.IssueValueNotAllowed}})
}

func TestPOSTDecoderRejectsEmptyBoolean(t *testing.T) {
	t.Parallel()
	document := []byte(`{"where":{"and":[]}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"/where/and", restqueryv1.IssueArityMismatch}})
}

func TestPOSTDecoderEnforcesBooleanPredicateUnion(t *testing.T) {
	t.Parallel()
	document := []byte(`{"where":{"and":[{"field":"id","op":"eq","value":"1"}],` +
		`"field":"id","op":"eq","value":"1"}}`)
	requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{
			{"/where/field", restqueryv1.IssueMemberForbidden},
			{"/where/op", restqueryv1.IssueMemberForbidden},
			{"/where/value", restqueryv1.IssueMemberForbidden},
		})
}
