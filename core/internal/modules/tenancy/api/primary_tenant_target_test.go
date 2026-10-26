package api

import (
	"errors"
	"testing"
)

const (
	targetEnvironmentID = "10000000-0000-4000-8000-000000000001"
	targetTenantID      = "50000000-0000-4000-8000-000000000001"
)

func TestPrimaryTenantTargetProjectsOnlyOnPremiseSlugPublicly(t *testing.T) {
	onPremise, err := NewPrimaryTenantTarget(
		targetEnvironmentID,
		DeploymentModeOnPremise,
		targetTenantID,
		PrimaryTenantPurposeOnPremise,
		"harpy",
	)
	if err != nil {
		t.Fatal(err)
	}
	if slug, visible := onPremise.PublicTenantSlug(); !visible || slug != "harpy" {
		t.Fatalf("on-premise public slug = %q, %t", slug, visible)
	}
	saas, err := NewPrimaryTenantTarget(
		targetEnvironmentID,
		DeploymentModeSaaS,
		targetTenantID,
		PrimaryTenantPurposeSaaS,
		"internal-control",
	)
	if err != nil {
		t.Fatal(err)
	}
	if slug, visible := saas.PublicTenantSlug(); visible || slug != "" {
		t.Fatalf("SaaS public slug = %q, %t", slug, visible)
	}
	if saas.CanonicalSlug() != "internal-control" {
		t.Fatalf("canonical slug = %q", saas.CanonicalSlug())
	}
}

func TestPrimaryTenantTargetRejectsInvalidBindings(t *testing.T) {
	tests := []struct {
		mode    DeploymentMode
		purpose PrimaryTenantPurpose
		slug    string
	}{
		{DeploymentModeOnPremise, PrimaryTenantPurposeSaaS, "harpy"},
		{DeploymentModeSaaS, PrimaryTenantPurposeOnPremise, "harpy"},
		{DeploymentModeOnPremise, PrimaryTenantPurposeOnPremise, " padded"},
	}
	for _, test := range tests {
		_, err := NewPrimaryTenantTarget(
			targetEnvironmentID,
			test.mode,
			targetTenantID,
			test.purpose,
			test.slug,
		)
		if !errors.Is(err, ErrPrimaryTenantTargetInvalid) {
			t.Fatalf("target %#v error = %v", test, err)
		}
	}
}
