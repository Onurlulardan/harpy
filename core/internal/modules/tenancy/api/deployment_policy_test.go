package api

import (
	"errors"
	"testing"
	"time"
)

const policyFixtureEnvironment = "10000000-0000-4000-8000-000000000001"

var policyFixtureTime = time.Date(2026, 7, 22, 9, 30, 0, 123000, time.UTC)

func TestNewDeploymentPolicyPreservesExactModeCapacity(t *testing.T) {
	zero := int64(0)
	onPremise, err := NewDeploymentPolicy(
		policyFixtureEnvironment, DeploymentModeOnPremise, &zero, policyFixtureTime,
	)
	if err != nil || !onPremise.IsValid() ||
		onPremise.EnvironmentID() != policyFixtureEnvironment ||
		onPremise.Mode() != DeploymentModeOnPremise ||
		!onPremise.CreatedAt().Equal(policyFixtureTime) {
		t.Fatalf("on-premise policy = %#v, %v", onPremise, err)
	}
	maximum, bounded := onPremise.MaximumCustomerTenants()
	if maximum != 0 || !bounded {
		t.Fatalf("on-premise capacity = %d, %t", maximum, bounded)
	}
	saas, err := NewDeploymentPolicy(
		policyFixtureEnvironment, DeploymentModeSaaS, nil, policyFixtureTime,
	)
	if err != nil || !saas.IsValid() {
		t.Fatalf("SaaS policy = %#v, %v", saas, err)
	}
	if maximum, bounded = saas.MaximumCustomerTenants(); maximum != 0 || bounded {
		t.Fatalf("SaaS capacity = %d, %t", maximum, bounded)
	}
}

func TestNewDeploymentPolicyRejectsInvalidAuthority(t *testing.T) {
	zero, one := int64(0), int64(1)
	tests := []struct {
		name        string
		environment string
		mode        DeploymentMode
		maximum     *int64
		createdAt   time.Time
	}{
		{"nil environment", "00000000-0000-0000-0000-000000000000", DeploymentModeOnPremise, &zero, policyFixtureTime},
		{"unknown mode", policyFixtureEnvironment, "SAAS", nil, policyFixtureTime},
		{"on-premise unbounded", policyFixtureEnvironment, DeploymentModeOnPremise, nil, policyFixtureTime},
		{"on-premise nonzero", policyFixtureEnvironment, DeploymentModeOnPremise, &one, policyFixtureTime},
		{"SaaS bounded", policyFixtureEnvironment, DeploymentModeSaaS, &zero, policyFixtureTime},
		{"zero time", policyFixtureEnvironment, DeploymentModeSaaS, nil, time.Time{}},
		{"non-UTC time", policyFixtureEnvironment, DeploymentModeSaaS, nil, policyFixtureTime.In(time.FixedZone("x", 60))},
		{"sub-microsecond time", policyFixtureEnvironment, DeploymentModeSaaS, nil, policyFixtureTime.Add(time.Nanosecond)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewDeploymentPolicy(
				test.environment, test.mode, test.maximum, test.createdAt,
			)
			if !errors.Is(err, ErrDeploymentPolicyInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
