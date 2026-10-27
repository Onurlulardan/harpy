package resolver

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveReturnsCanonicalFieldAndCapabilityIssues(t *testing.T) {
	fixture := newQueryFixture(t, true)
	page, _ := api.NewRequestPage(10, nil)
	unknown, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"missing"}, Page: page, Count: api.CountModeNone,
	})
	assertResolutionError(
		t, unknown, fixture.set, restqueryv1.ErrorQueryInvalid,
		"/select/0", restqueryv1.IssueFieldUnknown,
	)
	where, _ := api.NewRequestPredicate(
		"name", api.OperatorGT, []api.Literal{mustTextLiteral(t, "Ada")},
	)
	disallowed, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Where: &where, Page: page, Count: api.CountModeNone,
	})
	assertResolutionError(
		t, disallowed, fixture.set, restqueryv1.ErrorQueryInvalid,
		"/where/op", restqueryv1.IssueOperatorNotAllowed,
	)
}

func TestResolveReturnsCountAndIdentityBudgetErrors(t *testing.T) {
	disabled := newQueryFixture(t, false)
	page, _ := api.NewRequestPage(10, nil)
	exact, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Page: page, Count: api.CountModeExact,
	})
	assertResolutionError(
		t, exact, disabled.set, restqueryv1.ErrorQueryCountBudgetExceeded,
		"/count", restqueryv1.IssueCapabilityRequired,
	)
	enabled := newQueryFixture(t, true)
	large := mustTextLiteral(t, strings.Repeat("x", 65500))
	where, _ := api.NewRequestPredicate("name", api.OperatorEq, []api.Literal{large})
	request, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Where: &where, Page: page, Count: api.CountModeNone,
	})
	assertResolutionError(
		t, request, enabled.set, restqueryv1.ErrorQueryBudgetExceeded,
		"", restqueryv1.IssueQuerySizeExceeded,
	)
}

func TestResolveEnforcesExactCountBooleanBudget(t *testing.T) {
	fixture := newQueryFixture(t, true)
	filter, _ := api.NewRequestPredicate(
		"amount", api.OperatorEq, []api.Literal{mustTextLiteral(t, "1.20")},
	)
	for range 4 {
		filter, _ = api.NewRequestNot(filter)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Where: &filter, Page: page, Count: api.CountModeExact,
	})
	assertResolutionError(
		t, request, fixture.set, restqueryv1.ErrorQueryCountBudgetExceeded,
		"/where", restqueryv1.IssueBooleanDepthExceeded,
	)
}

func assertResolutionError(
	t *testing.T,
	request api.QueryRequest,
	catalogs api.CatalogSet,
	wantCode restqueryv1.ErrorCode,
	wantPath string,
	wantIssue restqueryv1.IssueCode,
) {
	t.Helper()
	_, err := Resolve(request, catalogs)
	queryError, ok := err.(api.QueryError)
	if !ok {
		t.Fatalf("error = %T %v, want api.QueryError", err, err)
	}
	issues := queryError.Issues()
	if queryError.Code() != wantCode || len(issues) != 1 ||
		issues[0].Path != wantPath || issues[0].Code != string(wantIssue) {
		t.Fatalf("code = %s, issues = %#v", queryError.Code(), issues)
	}
}
