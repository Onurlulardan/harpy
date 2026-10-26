package systemresourcequery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func EnsureGETSelect(base Request, mandatory []string) (Request, error) {
	value, err := base.Value()
	if err != nil || base.Transport() != "GET" ||
		value.Mode() != restquery.QueryModeRecord {
		return Request{}, ErrContract
	}
	selected := value.Select()
	if len(selected) == 0 {
		return base, nil
	}
	seen := make(map[string]struct{}, len(selected)+len(mandatory))
	for _, name := range selected {
		seen[name] = struct{}{}
	}
	for _, name := range mandatory {
		if name == "" {
			return Request{}, ErrContract
		}
		if _, found := seen[name]; found {
			continue
		}
		selected = append(selected, name)
		seen[name] = struct{}{}
	}
	if len(selected) > 50 {
		return Request{}, selectCountFailure()
	}
	input := restquery.RecordQueryRequestInput{
		Select: selected, Sort: value.Sort(), Page: value.Page(),
		Count: value.Count(), Include: value.Includes(),
	}
	if where, present := value.Where(); present {
		input.Where = &where
	}
	if search, present := value.Search(); present {
		input.Search = &search
	}
	rewritten, err := restquery.NewRecordQueryRequest(input)
	if err != nil {
		return Request{}, ErrContract
	}
	return newRequest(rewritten, "GET")
}

func selectCountFailure() error {
	issue, err := restquery.NewQueryIssue(
		"/select", restqueryv1.IssueSelectCountExceeded,
	)
	if err != nil {
		return ErrContract
	}
	queryErr, err := restquery.NewQueryError(
		restqueryv1.ErrorQueryBudgetExceeded, []runtimev1.Issue{issue},
	)
	if err != nil {
		return ErrContract
	}
	return NormalizeError(queryErr)
}
