package planner

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const (
	queryErrorBudget       = restqueryv1.ErrorQueryBudgetExceeded
	queryErrorCountBudget  = restqueryv1.ErrorQueryCountBudgetExceeded
	queryErrorTimeout      = restqueryv1.ErrorQueryTimeout
	queryErrorCountTimeout = restqueryv1.ErrorQueryCountTimeout
)

func budgetError(class StatementClass) error {
	code := queryErrorBudget
	if class == StatementClassExactCount {
		code = queryErrorCountBudget
	}
	return newQueryError(code, true)
}

func newQueryError(code restqueryv1.ErrorCode, plannerIssue bool) error {
	issues := []runtimev1.Issue{}
	if plannerIssue {
		issue, err := api.NewQueryIssue("", restqueryv1.IssuePlannerCostExceeded)
		if err != nil {
			return ErrPolicyInvalid
		}
		issues = append(issues, issue)
	}
	queryError, err := api.NewQueryError(code, issues)
	if err != nil {
		return ErrPolicyInvalid
	}
	return queryError
}
