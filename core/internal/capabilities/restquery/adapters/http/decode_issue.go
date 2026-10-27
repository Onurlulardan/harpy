package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type issueCollector struct {
	issues      []runtimev1.Issue
	ordinary    bool
	budget      bool
	countBudget bool
}

func (collector *issueCollector) add(path string, code restqueryv1.IssueCode) {
	collector.ordinary = true
	collector.append(path, code)
}

func (collector *issueCollector) addBudget(path string, code restqueryv1.IssueCode) {
	collector.budget = true
	collector.append(path, code)
}

func (collector *issueCollector) addCountBudget(path string, code restqueryv1.IssueCode) {
	collector.countBudget = true
	collector.append(path, code)
}

func (collector *issueCollector) append(path string, code restqueryv1.IssueCode) {
	issue, err := api.NewQueryIssue(path, code)
	if err != nil {
		issue, _ = api.NewQueryIssue("", restqueryv1.IssueValueNotCanonical)
		collector.ordinary = true
	}
	collector.issues = append(collector.issues, issue)
}

func (collector issueCollector) hasIssues() bool { return len(collector.issues) != 0 }

func (collector issueCollector) queryError() error {
	code := restqueryv1.ErrorQueryInvalid
	if !collector.ordinary && collector.countBudget {
		code = restqueryv1.ErrorQueryCountBudgetExceeded
	} else if !collector.ordinary && collector.budget {
		code = restqueryv1.ErrorQueryBudgetExceeded
	}
	return queryError(code, collector.issues)
}

func queryError(code restqueryv1.ErrorCode, issues []runtimev1.Issue) error {
	result, err := api.NewQueryError(code, issues)
	if err == nil {
		return result
	}
	fallback, _ := api.NewQueryError(restqueryv1.ErrorQueryInvalid, nil)
	return fallback
}

func singleQueryError(
	outer restqueryv1.ErrorCode, path string, code restqueryv1.IssueCode,
) error {
	issue, err := api.NewQueryIssue(path, code)
	if err != nil {
		return queryError(restqueryv1.ErrorQueryInvalid, nil)
	}
	return queryError(outer, []runtimev1.Issue{issue})
}
