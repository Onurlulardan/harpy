package httpquery

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderAllowsCombinedAggregateNodeBudgetBoundary(t *testing.T) {
	t.Parallel()
	document := aggregateNodeBudgetDocument(48, 2)
	decodeSuccess(t, document)
}

func TestPOSTDecoderRejectsCombinedAggregateNodeBudgetOverflow(t *testing.T) {
	t.Parallel()
	document := aggregateNodeBudgetDocument(48, 3)
	requireQueryError(t, []byte(document), restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/aggregate/having", restqueryv1.IssueASTNodeCountExceeded}})
}

func TestPOSTDecoderAllowsCombinedAggregateLiteralBudgetBoundary(t *testing.T) {
	t.Parallel()
	document := aggregateLiteralBudgetDocument(99)
	decodeSuccess(t, document)
}

func TestPOSTDecoderRejectsCombinedAggregateLiteralBudgetOverflow(t *testing.T) {
	t.Parallel()
	document := aggregateLiteralBudgetDocument(100)
	requireQueryError(t, []byte(document), restqueryv1.ErrorQueryBudgetExceeded,
		[]expectedQueryIssue{{"/aggregate/having", restqueryv1.IssueLiteralCountExceeded}})
}

func aggregateNodeBudgetDocument(whereCount, havingCount int) string {
	where := make([]string, whereCount)
	for index := range where {
		where[index] = `{"not":{"field":"status","op":"eq","value":"OPEN"}}`
	}
	having := make([]string, havingCount)
	for index := range having {
		having[index] = `{"measure":"rows","op":"eq","value":"1"}`
	}
	return `{"where":{"and":[` + strings.Join(where, ",") + `]},` +
		`"aggregate":{"measures":[{"name":"rows","function":"count"}],` +
		`"having":{"and":[` + strings.Join(having, ",") + `]}}}`
}

func aggregateLiteralBudgetDocument(lastWhereCount int) string {
	predicates := make([]string, 10)
	for index := range predicates {
		count := 100
		if index == len(predicates)-1 {
			count = lastWhereCount
		}
		predicates[index] = fmt.Sprintf(
			`{"field":"field_%d","op":"in","values":[%s]}`,
			index, quotedValues(count),
		)
	}
	return `{"where":{"and":[` + strings.Join(predicates, ",") + `]},` +
		`"aggregate":{"measures":[{"name":"rows","function":"count"}],` +
		`"having":{"measure":"rows","op":"eq","value":"1"}}}`
}
