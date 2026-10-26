package tenancy

import (
	"context"
	"fmt"
	"sync"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type DeploymentPolicyRuntime struct {
	expected tenancyapi.DeploymentMode
	query    DeploymentPolicyQuery
	targets  PrimaryTenantTargetQuery

	mu       sync.RWMutex
	policy   DeploymentPolicy
	target   PrimaryTenantTarget
	verified bool
}

func NewDeploymentPolicyRuntime(
	expected string,
	query DeploymentPolicyQuery,
	targets PrimaryTenantTargetQuery,
) (*DeploymentPolicyRuntime, error) {
	mode := tenancyapi.DeploymentMode(expected)
	if (mode != tenancyapi.DeploymentModeOnPremise &&
		mode != tenancyapi.DeploymentModeSaaS) || query == nil || targets == nil {
		return nil, ErrDeploymentPolicyDependencies
	}
	return &DeploymentPolicyRuntime{expected: mode, query: query, targets: targets}, nil
}

func (runtime *DeploymentPolicyRuntime) Verify(ctx context.Context) error {
	if runtime == nil || runtime.query == nil || runtime.targets == nil ||
		ctx == nil || ctx.Err() != nil {
		return ErrDeploymentPolicyDependencies
	}
	policy, err := runtime.query.GetDeploymentPolicy(ctx)
	if err != nil {
		runtime.clear()
		return fmt.Errorf("verify persisted deployment policy: %w", err)
	}
	if policy.Mode() != runtime.expected {
		runtime.clear()
		return ErrDeploymentModeMismatch
	}
	target, err := runtime.targets.GetPrimaryTenantTarget(ctx)
	if err != nil {
		runtime.clear()
		return fmt.Errorf("verify persisted primary tenant target: %w", err)
	}
	if target.EnvironmentID() != policy.EnvironmentID() ||
		target.DeploymentMode() != policy.Mode() {
		runtime.clear()
		return ErrDeploymentTargetMismatch
	}
	runtime.mu.Lock()
	runtime.policy = policy
	runtime.target = target
	runtime.verified = true
	runtime.mu.Unlock()
	return nil
}

func (runtime *DeploymentPolicyRuntime) PrimaryTarget() (PrimaryTenantTarget, error) {
	if runtime == nil {
		return PrimaryTenantTarget{}, ErrDeploymentPolicyNotVerified
	}
	runtime.mu.RLock()
	target, verified := runtime.target, runtime.verified
	runtime.mu.RUnlock()
	if !verified || !target.IsValid() {
		return PrimaryTenantTarget{}, ErrDeploymentPolicyNotVerified
	}
	return target, nil
}

func (runtime *DeploymentPolicyRuntime) Ready(ctx context.Context) bool {
	return runtime.Verify(ctx) == nil
}

func (runtime *DeploymentPolicyRuntime) Snapshot() (DeploymentPolicy, error) {
	if runtime == nil {
		return DeploymentPolicy{}, ErrDeploymentPolicyNotVerified
	}
	runtime.mu.RLock()
	policy, verified := runtime.policy, runtime.verified
	runtime.mu.RUnlock()
	if !verified || !policy.IsValid() {
		return DeploymentPolicy{}, ErrDeploymentPolicyNotVerified
	}
	return policy, nil
}

func (runtime *DeploymentPolicyRuntime) clear() {
	runtime.mu.Lock()
	runtime.policy = DeploymentPolicy{}
	runtime.target = PrimaryTenantTarget{}
	runtime.verified = false
	runtime.mu.Unlock()
}
