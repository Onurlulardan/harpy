package authorization

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type GlobalAuthorizer struct {
	owner     string
	port      GlobalPolicyPort
	validated bool
}

func NewGlobalAuthorizer(owner string, port GlobalPolicyPort) (GlobalAuthorizer, error) {
	authorizer := GlobalAuthorizer{owner: owner, port: port, validated: true}
	if authorizer.Validate() != nil {
		return GlobalAuthorizer{}, fmt.Errorf("REST Query global authorizer is invalid")
	}
	return authorizer, nil
}

func (authorizer GlobalAuthorizer) Validate() error {
	if !authorizer.validated || authorizer.owner == "" || globalPolicyPortNil(authorizer.port) {
		return fmt.Errorf("REST Query global authorizer is invalid")
	}
	return nil
}

func (authorizer GlobalAuthorizer) Authorize(
	ctx context.Context,
	scope transaction.GlobalScope,
	spec resolved.QuerySpec,
	snapshotAt time.Time,
) (resolved.AuthorizedPlan, error) {
	if ctx == nil || ctx.Err() != nil || authorizer.Validate() != nil ||
		transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireActiveGlobal(scope, transaction.SnapshotQuery) != nil ||
		transaction.RequireGlobalOwner(scope, authorizer.owner) != nil ||
		spec.Validate() != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	set, err := extractRequirements(spec)
	if err != nil || !requirementsOwnedBy(set, authorizer.owner) {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	identity, err := spec.IdentityHash()
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	request, err := newGlobalPolicyRequest(
		scope, identity, snapshotAt, authorizer.owner, set.requirements,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	decision, err := authorizer.port.EvaluateGlobalQuery(ctx, scope, request)
	if err != nil {
		return resolved.AuthorizedPlan{}, newPolicyEvaluationFailure(err)
	}
	if ctx.Err() != nil || transaction.RequireActiveGlobal(scope, transaction.SnapshotQuery) != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	validated, err := validateDecision(request.binding, request.requirements, decision)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyDecisionContract
	}
	return buildGlobalAuthorizedPlan(spec, set, request, validated)
}

func globalPolicyPortNil(port GlobalPolicyPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func requirementsOwnedBy(set requirementSet, owner string) bool {
	if !set.validate() || owner == "" {
		return false
	}
	for _, requirement := range set.requirements {
		current, err := requirement.catalog.Owner()
		if err != nil || current != owner {
			return false
		}
	}
	return true
}
