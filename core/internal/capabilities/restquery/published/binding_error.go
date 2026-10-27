package published

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type bindingIssueCollector struct {
	issues []runtimev1.Issue
}

func (collector *bindingIssueCollector) add(
	path string, code restqueryv1.IssueCode,
) {
	issue, err := api.NewQueryIssue(path, code)
	if err != nil {
		issue, _ = api.NewQueryIssue("", restqueryv1.IssueValueNotCanonical)
	}
	collector.issues = append(collector.issues, issue)
}

func (collector bindingIssueCollector) hasIssues() bool {
	return len(collector.issues) != 0
}

func (collector bindingIssueCollector) queryError() error {
	queryError, err := api.NewQueryError(
		restqueryv1.ErrorQueryParameterInvalid, collector.issues,
	)
	if err == nil {
		return queryError
	}
	fallback, _ := api.NewQueryError(restqueryv1.ErrorQueryParameterInvalid, nil)
	return fallback
}
