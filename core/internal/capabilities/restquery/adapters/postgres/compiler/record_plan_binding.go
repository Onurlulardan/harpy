package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func recordPlanBinding(
	plan resolved.AuthorizedPlan,
) (statementBinding, api.CountMode, error) {
	if plan.Validate() != nil {
		return statementBinding{}, 0, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	mode, err := spec.Mode()
	if err != nil || mode != api.QueryModeRecord {
		return statementBinding{}, 0, ErrContract
	}
	query, found, err := spec.Record()
	if err != nil || !found {
		return statementBinding{}, 0, ErrContract
	}
	count, err := recordShapeCount(query)
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	reference, referenceErr := spec.Reference()
	identity, identityErr := spec.IdentityHash()
	if referenceErr != nil || identityErr != nil {
		return statementBinding{}, 0, ErrContract
	}
	source, err := newCursorBinding(plan, reference, identity)
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	continuation, err := continuationIdentity(query)
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	binding, err := newStatementBinding(source, continuation)
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	return binding, count, nil
}
