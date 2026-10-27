package tenancy

import "errors"

var (
	ErrDeploymentPolicyDependencies = errors.New("deployment policy runtime dependencies are invalid")
	ErrDeploymentModeMismatch       = errors.New("persisted deployment policy does not match expected mode")
	ErrDeploymentTargetMismatch     = errors.New("primary tenant target does not match deployment policy")
	ErrDeploymentPolicyNotVerified  = errors.New("deployment policy runtime is not verified")
)
