package resolver

import (
	"fmt"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type resolutionIssues struct {
	code     restqueryv1.ErrorCode
	issues   []runtimev1.Issue
	internal error
}

func newResolutionIssues() *resolutionIssues {
	return &resolutionIssues{
		code:   restqueryv1.ErrorQueryInvalid,
		issues: []runtimev1.Issue{},
	}
}

func (collector *resolutionIssues) add(
	path string, code restqueryv1.IssueCode,
) {
	if collector.internal != nil {
		return
	}
	issue, err := api.NewQueryIssue(path, code)
	if err != nil {
		collector.internal = fmt.Errorf("REST Query resolver issue is invalid")
		return
	}
	collector.issues = append(collector.issues, issue)
}

func (collector *resolutionIssues) countBudget(
	path string, code restqueryv1.IssueCode,
) {
	collector.promote(restqueryv1.ErrorQueryCountBudgetExceeded)
	collector.add(path, code)
}

func (collector *resolutionIssues) budget(
	path string, code restqueryv1.IssueCode,
) {
	collector.promote(restqueryv1.ErrorQueryBudgetExceeded)
	collector.add(path, code)
}

func (collector *resolutionIssues) promote(code restqueryv1.ErrorCode) {
	if resolutionErrorPriority(code) > resolutionErrorPriority(collector.code) {
		collector.code = code
	}
}

func (collector *resolutionIssues) hasIssues() bool {
	return collector.internal != nil || len(collector.issues) != 0
}

func (collector *resolutionIssues) err() error {
	if collector.internal != nil {
		return collector.internal
	}
	queryError, err := api.NewQueryError(collector.code, collector.issues)
	if err != nil {
		return fmt.Errorf("REST Query resolver error is invalid")
	}
	return queryError
}

func resolutionErrorPriority(code restqueryv1.ErrorCode) int {
	switch code {
	case restqueryv1.ErrorQueryCountBudgetExceeded:
		return 3
	case restqueryv1.ErrorQueryBudgetExceeded:
		return 2
	default:
		return 1
	}
}

func invalidResolutionInput() error {
	collector := newResolutionIssues()
	collector.add("", restqueryv1.IssueReferenceInvalid)
	return collector.err()
}
