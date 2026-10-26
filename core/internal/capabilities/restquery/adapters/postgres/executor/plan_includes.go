package executor

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func planHasIncludes(plan resolved.AuthorizedPlan) (bool, error) {
	spec, err := plan.Spec()
	if err != nil {
		return false, ErrContract
	}
	record, found, err := spec.Record()
	if err != nil || !found {
		return false, ErrContract
	}
	includes, err := record.Includes()
	if err != nil {
		return false, ErrContract
	}
	return len(includes) > 0, nil
}
