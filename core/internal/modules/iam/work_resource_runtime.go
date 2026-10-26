package iam

import (
	"context"
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	workresourceapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workresource"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type WorkResourceRuntime struct {
	authorizer iamauthorization.Authorizer
	valid      bool
}

func NewWorkResourceRuntime() (*WorkResourceRuntime, error) {
	port, err := authorizationpostgres.NewWorkAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Work resource evidence: %w", err)
	}
	authorizer, err := iamauthorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM Work resource authorizer: %w", err)
	}
	return &WorkResourceRuntime{authorizer: authorizer, valid: true}, nil
}

func (runtime *WorkResourceRuntime) AuthorizeCurrent(
	ctx context.Context,
	scope platformtx.TenantScope,
	request workresourceapi.Request,
) (workresourceapi.Decision, error) {
	if runtime == nil || !runtime.valid || !request.Valid() ||
		scope.ActorID() == [16]byte{} || scope.TenantID() == [16]byte{} {
		return workresourceapi.Decision{}, workresourceapi.ErrInvalid
	}
	probe, err := iamauthorization.NewCurrentActorRequest(
		request.Resource, request.Action, scope.ActorID(),
	)
	if err != nil {
		return workresourceapi.Decision{}, workresourceapi.ErrInvalid
	}
	decision, err := runtime.authorizer.CurrentActorDecision(ctx, scope, probe)
	if err != nil || decision.Validate() != nil ||
		decision.AuthorizationVersion() < 1 || decision.AuthorizationClock() < 1 {
		return workresourceapi.Decision{}, workresourceapi.ErrUnavailable
	}
	return workresourceapi.Decision{Allowed: decision.Allowed(),
		AuthorizationRevision: decision.AuthorizationVersion(),
		AuthorizationClock:    decision.AuthorizationClock()}, nil
}

var _ workresourceapi.Authorizer = (*WorkResourceRuntime)(nil)
