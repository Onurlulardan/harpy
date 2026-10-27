package runtimev1

import (
	"strings"
	"testing"
)

func TestValidateBootstrapAuthorityManifestAcceptsExactOrigins(t *testing.T) {
	digest := strings.Repeat("a", 64)
	retained := validAuthorityManifest("v1_adopted", digest)
	authorityDatabase(retained)["extension_owner"] = "oracle"
	authorityDatabase(retained)["extension_owner_mode"] = "adopted_retained_inventory"
	authorityRole(retained, 11)["enabled"] = false
	fixtures := []map[string]any{
		validAuthorityManifest("fresh", digest),
		validAuthorityManifest("v1_adopted", digest),
		retained,
	}
	for index, fixture := range fixtures {
		manifest := typedContractFixture[BootstrapAuthorityManifest](t, fixture)
		if err := ValidateBootstrapAuthorityManifest(manifest); err != nil {
			t.Fatalf("valid authority fixture %d rejected: %v", index, err)
		}
	}
}

func TestValidateBootstrapAuthorityManifestRejectsSemanticDrift(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		mutate func(*BootstrapAuthorityManifest)
	}{
		{"environment identity", func(value *BootstrapAuthorityManifest) { value.EnvironmentID = "" }},
		{"incomplete role set", func(value *BootstrapAuthorityManifest) { value.Roles = value.Roles[:16] }},
		{"authority owner login", func(value *BootstrapAuthorityManifest) { value.Roles[1].Login = true }},
		{"adopter replication", func(value *BootstrapAuthorityManifest) { value.Roles[12].Replication = true }},
		{"decoder login", func(value *BootstrapAuthorityManifest) { value.Roles[13].Login = true }},
		{"adopter connection", func(value *BootstrapAuthorityManifest) { value.Roles[12].ConnectionLimit = 1 }},
		{"incomplete memberships", func(value *BootstrapAuthorityManifest) { value.Memberships = value.Memberships[:5] }},
		{"two leases", func(value *BootstrapAuthorityManifest) { value.CredentialLeases = value.CredentialLeases[:1] }},
		{"unique lease IDs", func(value *BootstrapAuthorityManifest) {
			value.CredentialLeases[1].LeaseID = value.CredentialLeases[0].LeaseID
		}},
		{"bounded lease", func(value *BootstrapAuthorityManifest) {
			value.CredentialLeases[0].ExpiresAt = "2026-07-13T00:16:00Z"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := typedContractFixture[BootstrapAuthorityManifest](
				t, validAuthorityManifest("fresh", digest),
			)
			test.mutate(&manifest)
			if err := ValidateBootstrapAuthorityManifest(manifest); err == nil {
				t.Fatal("unsafe bootstrap authority accepted")
			}
		})
	}
}

func TestValidateBootstrapAuthoritySignatureRejectsMalformedValue(t *testing.T) {
	signature := BootstrapAuthoritySignature{
		ContractVersion: "bootstrap-authority-signature.v1",
		ManifestSHA256:  strings.Repeat("a", 64),
		Signature: AttestationSignature{
			Algorithm: "Ed25519", KeyID: "release-1", Value: strings.Repeat("A", 86),
		},
	}
	if err := ValidateBootstrapAuthoritySignature(signature); err != nil {
		t.Fatalf("valid signature contract rejected: %v", err)
	}
	signature.Signature.Value = "not-base64url"
	if err := ValidateBootstrapAuthoritySignature(signature); err == nil {
		t.Fatal("malformed signature accepted")
	}
}
