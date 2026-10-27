package tenancy

import (
	"context"
	"errors"
	"testing"
	"time"
)

type policyQueryFixture struct {
	policy DeploymentPolicy
	err    error
}

type targetQueryFixture struct {
	target PrimaryTenantTarget
	err    error
}

func (query *targetQueryFixture) GetPrimaryTenantTarget(
	context.Context,
) (PrimaryTenantTarget, error) {
	return query.target, query.err
}

func (query *policyQueryFixture) GetDeploymentPolicy(
	context.Context,
) (DeploymentPolicy, error) {
	return query.policy, query.err
}

func TestDeploymentPolicyRuntimeVerifiesExactPersistedMode(t *testing.T) {
	t.Parallel()
	query := &policyQueryFixture{policy: policyFixture(t, DeploymentModeOnPremise)}
	targets := &targetQueryFixture{target: targetFixture(t, DeploymentModeOnPremise)}
	runtime, err := NewDeploymentPolicyRuntime("onpremise", query, targets)
	if err != nil || runtime.Verify(context.Background()) != nil {
		t.Fatalf("verify runtime = %v, %v", runtime, err)
	}
	if snapshot, snapshotErr := runtime.Snapshot(); snapshotErr != nil ||
		snapshot.Mode() != DeploymentModeOnPremise {
		t.Fatalf("snapshot = %#v, %v", snapshot, snapshotErr)
	}
	query.policy = policyFixture(t, DeploymentModeSaaS)
	if err := runtime.Verify(context.Background()); !errors.Is(err, ErrDeploymentModeMismatch) {
		t.Fatalf("mismatch error = %v", err)
	}
	if _, err := runtime.Snapshot(); !errors.Is(err, ErrDeploymentPolicyNotVerified) {
		t.Fatalf("stale snapshot remained available: %v", err)
	}
}

func TestDeploymentPolicyRuntimeFailsClosedWhenPolicyUnavailable(t *testing.T) {
	t.Parallel()
	query := &policyQueryFixture{policy: policyFixture(t, DeploymentModeSaaS)}
	targets := &targetQueryFixture{target: targetFixture(t, DeploymentModeSaaS)}
	runtime, _ := NewDeploymentPolicyRuntime("saas", query, targets)
	if err := runtime.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	query.err = errors.New("policy unavailable")
	if runtime.Ready(context.Background()) {
		t.Fatal("unavailable policy reported ready")
	}
	if _, err := runtime.Snapshot(); !errors.Is(err, ErrDeploymentPolicyNotVerified) {
		t.Fatalf("unavailable snapshot error = %v", err)
	}
}

func TestDeploymentPolicyRuntimeFailsClosedWithoutPrimaryTarget(t *testing.T) {
	t.Parallel()
	query := &policyQueryFixture{policy: policyFixture(t, DeploymentModeOnPremise)}
	targets := &targetQueryFixture{err: errors.New("target unavailable")}
	runtime, _ := NewDeploymentPolicyRuntime("onpremise", query, targets)
	if runtime.Ready(context.Background()) {
		t.Fatal("missing primary target reported ready")
	}
	if _, err := runtime.PrimaryTarget(); !errors.Is(err, ErrDeploymentPolicyNotVerified) {
		t.Fatalf("missing target snapshot error = %v", err)
	}
}

func policyFixture(t *testing.T, mode DeploymentMode) DeploymentPolicy {
	t.Helper()
	var maximum *int64
	if mode == DeploymentModeOnPremise {
		zero := int64(0)
		maximum = &zero
	}
	policy, err := NewDeploymentPolicy(
		"10000000-0000-4000-8000-000000000001",
		mode,
		maximum,
		time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func targetFixture(t *testing.T, mode DeploymentMode) PrimaryTenantTarget {
	t.Helper()
	purpose, slug := PrimaryTenantPurposeOnPremise, "harpy"
	if mode == DeploymentModeSaaS {
		purpose, slug = PrimaryTenantPurposeSaaS, "control"
	}
	target, err := NewPrimaryTenantTarget(
		"10000000-0000-4000-8000-000000000001",
		mode,
		"20000000-0000-4000-8000-000000000001",
		purpose,
		slug,
	)
	if err != nil {
		t.Fatal(err)
	}
	return target
}
