package executor

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func includeResultBudgetError() error {
	issue, err := api.NewQueryIssue("", restqueryv1.IssueIncludeResultCountExceeded)
	if err != nil {
		return ErrContract
	}
	result, err := api.NewQueryError(
		restqueryv1.ErrorQueryBudgetExceeded, []runtimev1.Issue{issue},
	)
	if err != nil {
		return ErrContract
	}
	return result
}

func relationCardinalityError() error {
	result, err := api.NewQueryError(
		restqueryv1.ErrorRelationCardinalityViolation, nil,
	)
	if err != nil {
		return ErrContract
	}
	return result
}
