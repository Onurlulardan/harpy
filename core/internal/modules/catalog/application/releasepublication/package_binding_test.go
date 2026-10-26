package releasepublication

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	iamcatalogpayload "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/catalogpayload"
)

func TestVerifyOwnerPayloadBindsSourceContractToPackageSlug(t *testing.T) {
	const componentID = "018f0000-0005-7000-8000-000000000001"
	value := iamcatalogpayload.Payload{
		ContractVersion: catalogpackagev1.ContractIAM,
		ComponentID:     componentID,
		ModuleSlug:      "sales-crm",
		SourceContract:  "harpy-v1.enterprise-seed.sales-crm",
		Resources: []iamcatalogpayload.Resource{
			{
				Slug: "module.sales-crm", Kind: "module",
				Name: "Sales CRM", Description: "Module",
			},
			{
				Slug: "form.crm-customer", Kind: "form",
				Name: "Customer", Description: "Form",
			},
		},
		Grants: []iamcatalogpayload.Grant{
			{
				ResourceSlug: "module.sales-crm", RoleName: "Admin",
				RoleDescription: "Administrator", Actions: []string{"read"},
			},
			{
				ResourceSlug: "form.crm-customer", RoleName: "Admin",
				RoleDescription: "Administrator", Actions: []string{"read", "write"},
			},
		},
	}
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	document, err = canonicaljsonv1.CanonicalizeObject(
		document, iamcatalogpayload.MaximumPayloadBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	payload := catalogpackagev1.OwnerPayload{
		ComponentID:   componentID,
		Owner:         "iam",
		Contract:      catalogpackagev1.ContractIAM,
		ContentSHA256: hex.EncodeToString(digest[:]),
		Document:      document,
	}
	if err := verifyOwnerPayload(
		"sales-crm", payload.Owner, payload.Contract, payload,
	); err != nil {
		t.Fatalf("exact package source binding was rejected: %v", err)
	}
	if err := verifyOwnerPayload(
		"crm", payload.Owner, payload.Contract, payload,
	); err == nil {
		t.Fatal("payload source contract differing from package slug was accepted")
	}
}
