package api

import (
	"cmp"
	"fmt"
	"io"
	"slices"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const maximumQueryIssues = 20

type QueryError struct {
	code   restqueryv1.ErrorCode
	issues []runtimev1.Issue
}

func NewQueryError(
	code restqueryv1.ErrorCode, issues []runtimev1.Issue,
) (QueryError, error) {
	status, found := restqueryv1.ErrorStatus(code)
	if !code.Valid() || !found || status < 400 || status > 599 {
		return QueryError{}, fmt.Errorf("REST Query error code is invalid")
	}
	owned := append([]runtimev1.Issue(nil), issues...)
	for index, issue := range owned {
		if err := issue.Validate(); err != nil {
			return QueryError{}, fmt.Errorf("REST Query issue %d is invalid: %w", index, err)
		}
		if !restqueryv1.IssueCode(issue.Code).Valid() {
			return QueryError{}, fmt.Errorf("REST Query issue %d code is not published", index)
		}
	}
	sortQueryIssues(owned)
	owned = uniqueQueryIssues(owned)
	if len(owned) > maximumQueryIssues {
		owned = owned[:maximumQueryIssues]
	}
	owned = append([]runtimev1.Issue{}, owned...)
	return QueryError{code: code, issues: owned}, nil
}

func NewQueryIssue(
	path string, code restqueryv1.IssueCode,
) (runtimev1.Issue, error) {
	if !code.Valid() {
		return runtimev1.Issue{}, fmt.Errorf("REST Query issue code is invalid")
	}
	return runtimev1.NewIssue(path, string(code))
}

func (queryError QueryError) Code() restqueryv1.ErrorCode { return queryError.code }

func (queryError QueryError) HTTPStatus() (int, bool) {
	return restqueryv1.ErrorStatus(queryError.code)
}

func (queryError QueryError) Issues() []runtimev1.Issue {
	return append([]runtimev1.Issue{}, queryError.issues...)
}

func (queryError QueryError) Error() string    { return string(queryError.code) }
func (queryError QueryError) String() string   { return string(queryError.code) }
func (queryError QueryError) GoString() string { return string(queryError.code) }
func (queryError QueryError) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, string(queryError.code))
}
func (QueryError) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func sortQueryIssues(issues []runtimev1.Issue) {
	slices.SortFunc(issues, func(left, right runtimev1.Issue) int {
		if order := cmp.Compare(left.Path, right.Path); order != 0 {
			return order
		}
		return cmp.Compare(left.Code, right.Code)
	})
}

func uniqueQueryIssues(issues []runtimev1.Issue) []runtimev1.Issue {
	write := 0
	for _, issue := range issues {
		if write > 0 && issues[write-1] == issue {
			continue
		}
		issues[write] = issue
		write++
	}
	return issues[:write]
}
