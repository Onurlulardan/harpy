package authorization

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func queryErrorForDenials(
	denied []Requirement,
) (api.QueryError, bool, error) {
	if len(denied) == 0 {
		return api.QueryError{}, false, nil
	}
	issues := make([]runtimev1.Issue, 0, len(denied))
	for _, requirement := range denied {
		if requirement.Validate() != nil {
			return api.QueryError{}, false, errPolicyEvaluationContract
		}
		kind, _ := requirement.Kind()
		if !fieldRequirement(kind) {
			queryError, err := api.NewQueryError(
				restqueryv1.ErrorQueryForbidden, nil,
			)
			return queryError, true, err
		}
		path, err := requirement.RequestPath()
		if err != nil {
			return api.QueryError{}, false, errPolicyEvaluationContract
		}
		value, err := path.Value()
		if err != nil {
			return api.QueryError{}, false, errPolicyEvaluationContract
		}
		issue, err := api.NewQueryIssue(
			value, restqueryv1.IssueFieldForbidden,
		)
		if err != nil {
			return api.QueryError{}, false, errPolicyEvaluationContract
		}
		issues = append(issues, issue)
	}
	return fieldDenialError(issues)
}

func fieldDenialError(
	issues []runtimev1.Issue,
) (api.QueryError, bool, error) {
	queryError, err := api.NewQueryError(
		restqueryv1.ErrorQueryFieldForbidden, issues,
	)
	return queryError, true, err
}

func fieldRequirement(kind RequirementKind) bool {
	switch kind {
	case RequirementFieldSelect, RequirementFieldFilter,
		RequirementFieldSort, RequirementFieldSearch,
		RequirementFieldGroup, RequirementFieldMeasure,
		RequirementFieldHaving:
		return true
	default:
		return false
	}
}
