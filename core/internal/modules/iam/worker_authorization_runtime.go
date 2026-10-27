package iam

import (
	"context"
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type WorkerAuthorizationRuntime struct {
	authorizer iamauthorization.Authorizer
}

func NewAnalyticsWorkerAuthorizationRuntime() (*WorkerAuthorizationRuntime, error) {
	evidence, err := authorizationpostgres.NewAnalyticsAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build Analytics IAM worker evidence: %w", err)
	}
	return newWorkerAuthorizationRuntime(evidence)
}

func NewIAMWorkerAuthorizationRuntime() (*WorkerAuthorizationRuntime, error) {
	evidence, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM worker evidence: %w", err)
	}
	return newWorkerAuthorizationRuntime(evidence)
}

func newWorkerAuthorizationRuntime(
	evidence iamauthorization.EvidencePort,
) (*WorkerAuthorizationRuntime, error) {
	authorizer, err := iamauthorization.NewAuthorizer(evidence)
	if err != nil {
		return nil, fmt.Errorf("build IAM worker authorizer: %w", err)
	}
	return &WorkerAuthorizationRuntime{authorizer: authorizer}, nil
}

func (runtime *WorkerAuthorizationRuntime) Authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource string,
	action string,
	tokenVersion int64,
	authorizationVersion int64,
) (bool, uuid.UUID, int64, int64, int64, error) {
	request, err := iamauthorization.NewRequest(
		resource, action, tokenVersion, authorizationVersion,
	)
	if err != nil || runtime == nil {
		return false, uuid.Nil, 0, 0, 0, fmt.Errorf("IAM worker request is invalid")
	}
	decision, err := runtime.authorizer.Authorize(ctx, scope, request)
	if err != nil {
		return false, uuid.Nil, 0, 0, 0, err
	}
	permissionID, _ := decision.PermissionID()
	return decision.Allowed(), permissionID, tokenVersion,
		decision.AuthorizationVersion(), decision.AuthorizationClock(), nil
}

func (runtime *WorkerAuthorizationRuntime) AuthorizePinned(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource string,
	action string,
	authorizationVersion int64,
) (bool, uuid.UUID, int64, int64, int64, error) {
	request, err := iamauthorization.NewPinnedActorRequest(
		resource, action, authorizationVersion,
	)
	if err != nil || runtime == nil {
		return false, uuid.Nil, 0, 0, 0, fmt.Errorf("IAM pinned worker request is invalid")
	}
	decision, err := runtime.authorizer.ResolvePinnedActor(ctx, scope, request)
	if err != nil {
		return false, uuid.Nil, 0, 0, 0, err
	}
	permissionID, _ := decision.PermissionID()
	return decision.Allowed(), permissionID, decision.TokenVersion(),
		decision.AuthorizationVersion(), decision.AuthorizationClock(), nil
}
