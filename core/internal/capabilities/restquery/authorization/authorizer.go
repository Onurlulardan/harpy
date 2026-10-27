package authorization

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Authorizer struct {
	registry  PolicyRegistry
	validated bool
}

func NewAuthorizer(registry PolicyRegistry) (Authorizer, error) {
	authorizer := Authorizer{registry: registry, validated: true}
	if err := authorizer.Validate(); err != nil {
		return Authorizer{}, err
	}
	return authorizer, nil
}

func (authorizer Authorizer) Validate() error {
	if !authorizer.validated || authorizer.registry.Validate() != nil {
		return fmt.Errorf("REST Query authorizer is invalid")
	}
	return nil
}

func (authorizer Authorizer) Authorize(
	ctx context.Context,
	scope transaction.TenantScope,
	spec resolved.QuerySpec,
) (resolved.AuthorizedPlan, error) {
	if ctx == nil || authorizer.Validate() != nil ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil ||
		transaction.NetworkAllowed(ctx) || ctx.Err() != nil || spec.Validate() != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	set, err := extractRequirements(spec)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	queryIdentity, err := spec.IdentityHash()
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	requests, err := groupPolicyRequests(
		scope.TenantID(), scope.ActorID(), queryIdentity, set.requirements,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	evaluations, err := evaluatePolicyRequests(
		ctx, scope, authorizer.registry, requests,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, err
	}
	return buildAuthorizedPlan(
		spec, set, scope.TenantID(), scope.ActorID(), evaluations,
	)
}

func (Authorizer) String() string   { return authorizationRedaction }
func (Authorizer) GoString() string { return authorizationRedaction }
func (Authorizer) LogValue() slog.Value {
	return authorizationLogValue()
}
func (authorizer Authorizer) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (Authorizer) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
