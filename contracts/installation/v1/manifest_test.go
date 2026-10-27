package installationv1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestParseAcceptsStrictImmutableManifest(t *testing.T) {
	document := fixtureManifest(t, DeploymentOnPremise)
	manifest, err := Parse(document)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.EnvironmentIDString() != "018f0000-0000-7000-8000-000000000001" ||
		manifest.DeploymentMode() != DeploymentOnPremise ||
		manifest.PrimaryTenant().Slug != "harpy" || len(manifest.CanonicalBytes()) == 0 {
		t.Fatal("parsed installation manifest drifted")
	}
	digest := sha256.Sum256(manifest.CanonicalBytes())
	if manifest.SHA256() != hex.EncodeToString(digest[:]) {
		t.Fatal("manifest digest does not bind canonical bytes")
	}
}

func TestParseRejectsUnknownDuplicateSecretAndHashDrift(t *testing.T) {
	valid := string(fixtureManifest(t, DeploymentSaaS))
	tests := []string{
		strings.Replace(valid, `"request_sha256":`, `"unknown":true,"request_sha256":`, 1),
		strings.Replace(valid, `"environment_id":`, `"environment_id":"018f0000-0000-7000-8000-000000000001","environment_id":`, 1),
		strings.Replace(valid, `"request_sha256":"`, `"password":"secret","request_sha256":"`, 1),
		strings.Replace(valid, `"request_sha256":"`, `"request_sha256":"0`, 1),
		strings.Replace(valid, `"deployment_mode":"saas"`, `"deployment_mode":"SAAS"`, 1),
	}
	for _, document := range tests {
		if _, err := Parse([]byte(document)); err == nil {
			t.Fatalf("invalid installation manifest accepted: %s", document)
		}
	}
}

func fixtureManifest(t *testing.T, mode DeploymentMode) []byte {
	t.Helper()
	wire := manifestWire{ContractVersion: ContractVersion,
		EnvironmentID: "018f0000-0000-7000-8000-000000000001", DeploymentMode: mode,
		PrimaryTenant: TenantMetadata{Name: "Harpy", Slug: "harpy"},
		FirstAdmin:    AdminMetadata{DisplayName: "Instance Owner", Email: "owner@harpy.io"}}
	wire.RequestSHA256 = requestDigest(wire)
	return []byte(fmt.Sprintf(`{"contract_version":"%s","environment_id":"%s","deployment_mode":"%s","primary_tenant":{"name":"%s","slug":"%s"},"first_admin":{"display_name":"%s","email":"%s"},"request_sha256":"%s"}`,
		wire.ContractVersion, wire.EnvironmentID, wire.DeploymentMode,
		wire.PrimaryTenant.Name, wire.PrimaryTenant.Slug,
		wire.FirstAdmin.DisplayName, wire.FirstAdmin.Email, wire.RequestSHA256))
}
