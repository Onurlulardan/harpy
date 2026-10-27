package authorization

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Authorizer struct {
	port  EvidencePort
	valid bool
}

func NewAuthorizer(port EvidencePort) (Authorizer, error) {
	if evidencePortIsNil(port) {
		return Authorizer{}, ErrAuthorizationContract
	}
	return Authorizer{port: port, valid: true}, nil
}

func (authorizer Authorizer) Authorize(
	ctx context.Context,
	scope transaction.TenantScope,
	request Request,
) (Decision, error) {
	if !authorizer.valid || evidencePortIsNil(authorizer.port) ||
		request.Validate() != nil || !validAuthorizationScope(ctx, scope) {
		return deniedDecision(ReasonInvalidRequest, 0, 0), ErrAuthorizationContract
	}
	snapshot, err := authorizer.port.Authorize(ctx, scope, request)
	if err != nil {
		return deniedDecision(ReasonEvidenceUnavailable, 0, 0),
			fmt.Errorf("%w: %w", ErrAuthorizationEvidence, err)
	}
	if validateErr := snapshot.Validate(); validateErr != nil {
		return deniedDecision(ReasonEvidenceUnavailable, 0, 0),
			fmt.Errorf("%w: %w", ErrAuthorizationEvidence, validateErr)
	}
	return evaluate(request, snapshot), nil
}

func validAuthorizationScope(
	ctx context.Context,
	scope transaction.TenantScope,
) bool {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsScope(ctx, scope) {
		return false
	}
	switch scope.Mode() {
	case transaction.Command:
		return transaction.RequireActive(scope, transaction.Command) == nil
	case transaction.SnapshotQuery:
		return transaction.RequireActive(scope, transaction.SnapshotQuery) == nil
	default:
		return false
	}
}
