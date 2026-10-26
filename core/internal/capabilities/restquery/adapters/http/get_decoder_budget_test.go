package httpquery

import (
	"strconv"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestGETDecoderEnforcesRecordBudgets(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"select", getSelectQuery(51), "/select", restqueryv1.IssueSelectCountExceeded},
		{"sort", getSortQuery(6), "/sort", restqueryv1.IssueSortCountExceeded},
		{"filter", getFilterQuery(51), "/filter", restqueryv1.IssuePredicateCountExceeded},
		{"operand", getOperandQuery(101), "/filter/0/values", restqueryv1.IssueOperandCountExceeded},
		{"search runes", "search[term]=" + strings.Repeat("a", 129),
			"/search/term", restqueryv1.IssueSearchSizeExceeded},
		{"search bytes", "search[term]=" + strings.Repeat("€", 86),
			"/search/term", restqueryv1.IssueSearchSizeExceeded},
		{"page", "page[size]=501", "/page/size", restqueryv1.IssuePageSizeExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireGETError(t, test.raw, restqueryv1.ErrorQueryBudgetExceeded,
				[]expectedGETIssue{{test.path, test.issue}})
		})
	}
}

func TestGETDecoderUsesCountSpecificBudget(t *testing.T) {
	rawQuery := getFilterQuery(21) + "&count=exact"
	requireGETError(t, rawQuery, restqueryv1.ErrorQueryCountBudgetExceeded,
		[]expectedGETIssue{{"/where", restqueryv1.IssuePredicateCountExceeded}})
}

func getSelectQuery(count int) string {
	leaves := make([]string, count)
	for index := range count {
		leaves[index] = "select[" + strconv.Itoa(index) + "]=f" + strconv.Itoa(index)
	}
	return strings.Join(leaves, "&")
}

func getSortQuery(count int) string {
	leaves := make([]string, 0, count*2)
	for index := range count {
		prefix := "sort[" + strconv.Itoa(index) + "]"
		leaves = append(leaves, prefix+"[field]=f"+strconv.Itoa(index),
			prefix+"[direction]=asc")
	}
	return strings.Join(leaves, "&")
}

func getFilterQuery(count int) string {
	leaves := make([]string, 0, count*3)
	for index := range count {
		prefix := "filter[" + strconv.Itoa(index) + "]"
		leaves = append(leaves, prefix+"[field]=f"+strconv.Itoa(index),
			prefix+"[op]=eq", prefix+"[value]=x")
	}
	return strings.Join(leaves, "&")
}

func getOperandQuery(count int) string {
	leaves := []string{"filter[0][field]=f", "filter[0][op]=in"}
	for index := range count {
		leaves = append(leaves,
			"filter[0][values]["+strconv.Itoa(index)+"]=x")
	}
	return strings.Join(leaves, "&")
}
