package environmentoperator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func operatorManifest(t *testing.T, mode installationv1.DeploymentMode) installationv1.Manifest {
	t.Helper()
	tenant := installationv1.TenantMetadata{Name: "ForgeCRUD", Slug: "forgecrud"}
	admin := installationv1.AdminMetadata{DisplayName: "Platform Owner",
		Email: "owner@forgecrud.test"}
	request := struct {
		EnvironmentID  string                        `json:"environment_id"`
		DeploymentMode installationv1.DeploymentMode `json:"deployment_mode"`
		PrimaryTenant  installationv1.TenantMetadata `json:"primary_tenant"`
		FirstAdmin     installationv1.AdminMetadata  `json:"first_admin"`
	}{"10000000-0000-4000-8000-000000000001", mode, tenant, admin}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 32*1024)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	document := struct {
		ContractVersion string                        `json:"contract_version"`
		EnvironmentID   string                        `json:"environment_id"`
		DeploymentMode  installationv1.DeploymentMode `json:"deployment_mode"`
		PrimaryTenant   installationv1.TenantMetadata `json:"primary_tenant"`
		FirstAdmin      installationv1.AdminMetadata  `json:"first_admin"`
		RequestSHA256   string                        `json:"request_sha256"`
	}{installationv1.ContractVersion, request.EnvironmentID, mode, tenant, admin,
		hex.EncodeToString(digest[:])}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := installationv1.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}
