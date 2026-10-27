package httpquery

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderUsesPublishedRootLimitIssueCodes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		document string
		path     string
		issue    restqueryv1.IssueCode
	}{
		{"page", `{"page":{"size":501,"after":null}}`, "/page/size", restqueryv1.IssuePageSizeExceeded},
		{"aggregate", `{"aggregate":{"limit":501}}`, "/aggregate/limit", restqueryv1.IssueResultCountExceeded},
		{"select", `{"select":[` + quotedNames(51) + `]}`, "/select", restqueryv1.IssueSelectCountExceeded},
		{"sort", `{"sort":[` + sortEntries(6) + `]}`, "/sort", restqueryv1.IssueSortCountExceeded},
		{"operand", `{"where":{"field":"id","op":"in","values":[` +
			quotedValues(101) + `]}}`, "/where/values", restqueryv1.IssueOperandCountExceeded},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			requireQueryError(t, []byte(test.document), restqueryv1.ErrorQueryBudgetExceeded,
				[]expectedQueryIssue{{test.path, test.issue}})
		})
	}
}

func TestPOSTDecoderUsesCountSpecificBudgetError(t *testing.T) {
	t.Parallel()
	leaf := `{"field":"status","op":"eq","value":"OPEN"}`
	deep := strings.Repeat(`{"not":`, 4) + leaf + strings.Repeat(`}`, 4)
	requireQueryError(t, []byte(`{"count":"exact","where":`+deep+`}`),
		restqueryv1.ErrorQueryCountBudgetExceeded,
		[]expectedQueryIssue{{"/where", restqueryv1.IssueBooleanDepthExceeded}})

	predicates := make([]string, 21)
	for index := range predicates {
		predicates[index] = leaf
	}
	document := `{"count":"exact","where":{"and":[` + strings.Join(predicates, ",") + `]}}`
	requireQueryError(t, []byte(document), restqueryv1.ErrorQueryCountBudgetExceeded,
		[]expectedQueryIssue{{"/where", restqueryv1.IssuePredicateCountExceeded}})
}

func quotedNames(count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = fmt.Sprintf("%q", fmt.Sprintf("field_%d", index))
	}
	return strings.Join(values, ",")
}

func quotedValues(count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = `"x"`
	}
	return strings.Join(values, ",")
}

func sortEntries(count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = fmt.Sprintf(`{"field":"field_%d","direction":"asc"}`, index)
	}
	return strings.Join(values, ",")
}
