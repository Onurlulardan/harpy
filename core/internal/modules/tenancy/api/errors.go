package api

import "errors"

var (
	ErrDeploymentPolicyInvalid     = errors.New("tenancy deployment policy is invalid")
	ErrDeploymentPolicyMissing     = errors.New("tenancy deployment policy is missing")
	ErrDeploymentPolicyConflict    = errors.New("tenancy deployment policy conflicts")
	ErrDeploymentPolicyUnavailable = errors.New(
		"tenancy deployment policy is unavailable",
	)
	ErrPrimaryTenantTargetInvalid = errors.New(
		"tenancy primary tenant target is invalid",
	)
	ErrPrimaryTenantTargetMissing = errors.New(
		"tenancy primary tenant target is missing",
	)
	ErrPrimaryTenantTargetUnavailable = errors.New(
		"tenancy primary tenant target is unavailable",
	)
)
