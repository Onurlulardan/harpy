package catalogpayload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

func TestVerifyAcceptsPayloadDerivedFirstPartyIAMComposition(t *testing.T) {
	componentID := "018f0000-0003-7000-8000-000000000009"
	payload := Payload{
		ContractVersion: catalogpackagev1.ContractIAM,
		ComponentID:     componentID, ModuleSlug: "project",
		SourceContract: "harpy-v1.enterprise-seed.project",
		Resources: []Resource{
			{Slug: "module.project", Kind: "module", Name: "Project", Description: "Module"},
			{Slug: "form.project-task", Kind: "form", Name: "Task", Description: "Form"},
		},
		Grants: []Grant{
			{ResourceSlug: "module.project", RoleName: "Admin",
				RoleDescription: "Administrator", Actions: []string{"read"}},
			{ResourceSlug: "form.project-task", RoleName: "Admin",
				RoleDescription: "Administrator", Actions: []string{"read", "write"}},
		},
	}
	document := canonicalIAMJSON(t, payload)
	digest := sha256.Sum256(document)
	verified, err := Verify(document, componentID, hex.EncodeToString(digest[:]))
	if err != nil || len(verified.Value().Resources) != 2 {
		t.Fatalf("verify project IAM: value=%#v err=%v", verified.Value(), err)
	}
	payload.Resources[0].Slug = "module.hr"
	mismatched := canonicalIAMJSON(t, payload)
	mismatchedDigest := sha256.Sum256(mismatched)
	if _, err = Verify(
		mismatched, componentID, hex.EncodeToString(mismatchedDigest[:]),
	); err == nil {
		t.Fatal("IAM accepted a module resource that differs from module_slug")
	}
}

func canonicalIAMJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
