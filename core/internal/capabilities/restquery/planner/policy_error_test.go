package planner

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestStatementTimeoutErrorMapping(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		class StatementClass
		code  restqueryv1.ErrorCode
	}{
		{StatementClassRecord, restqueryv1.ErrorQueryTimeout},
		{StatementClassInclude, restqueryv1.ErrorQueryTimeout},
		{StatementClassV1PredecessorSeek, restqueryv1.ErrorQueryTimeout},
		{StatementClassExactCount, restqueryv1.ErrorQueryCountTimeout},
		{StatementClassAggregate, restqueryv1.ErrorQueryTimeout},
	} {
		statement, _ := HardPolicy().Statement(test.class)
		requireQueryError(t, statement.TimeoutError(), test.code, false)
	}
}

func requireQueryError(
	t *testing.T, err error, code restqueryv1.ErrorCode, plannerIssue bool,
) api.QueryError {
	t.Helper()
	var queryError api.QueryError
	if !errors.As(err, &queryError) || queryError.Code() != code {
		t.Fatalf("query error = %T %v, want %s", err, err, code)
	}
	issues := queryError.Issues()
	if !plannerIssue && len(issues) == 0 {
		return queryError
	}
	if !plannerIssue || len(issues) != 1 || issues[0].Path != "" ||
		issues[0].Code != string(restqueryv1.IssuePlannerCostExceeded) {
		t.Fatalf("query issues = %#v", issues)
	}
	return queryError
}
