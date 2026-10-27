package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

var aggregateContinuation = sha256.Sum256(
	[]byte("forgecrud.restquery.postgres.aggregate-continuation.v1"),
)

func aggregatePlanBinding(
	plan resolved.AuthorizedPlan,
) (statementBinding, error) {
	if plan.Validate() != nil {
		return statementBinding{}, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return statementBinding{}, ErrContract
	}
	mode, err := spec.Mode()
	if err != nil || mode != api.QueryModeAggregate {
		return statementBinding{}, ErrContract
	}
	query, found, err := spec.Aggregate()
	if err != nil || !found || query.Validate() != nil {
		return statementBinding{}, ErrContract
	}
	reference, referenceErr := spec.Reference()
	identity, identityErr := spec.IdentityHash()
	if referenceErr != nil || identityErr != nil {
		return statementBinding{}, ErrContract
	}
	source, err := newCursorBinding(plan, reference, identity)
	if err != nil {
		return statementBinding{}, ErrContract
	}
	binding, err := newStatementBinding(source, aggregateContinuation)
	if err != nil {
		return statementBinding{}, ErrContract
	}
	return binding, nil
}
