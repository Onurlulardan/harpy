//go:build integration

package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const catalogDemoSalesPackageID = "e3000000-0000-0000-0000-000000000003"
const catalogDemoSalesReleaseID = "019c3300-0000-7000-8000-000000000003"
const catalogDemoSalesComponentID = "018f0000-0004-7000-8000-000000000003"
const catalogDemoLostReasonID = "018f0000-0000-7000-8000-000000000080"
const catalogDemoLostReasonSHA = "ec863428496e286c6668b7fe9a797d3c37714bdc10363191216973f740b53c05"
const catalogDemoCrossLegacy = "a1000000-0000-5000-8000-000000000001"

func catalogDemoSalesFormsRequest(t *testing.T, root string, owner any) any {
	t.Helper()
	path := filepath.Join(root, "tools", "internal", "firstpartycatalog",
		"testdata", "sales-crm", "artifacts", "owner-forms.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var group catalogDemoOwnerPayloadGroup
	if err = json.Unmarshal(content, &group); err != nil {
		t.Fatalf("decode Sales Forms owner artifact: %v", err)
	}
	for _, payload := range group.Payloads {
		var identity struct {
			FormID string `json:"form_id"`
			Slug   string `json:"slug"`
		}
		if json.Unmarshal(payload.Document, &identity) != nil ||
			identity.Slug != "crm-lost-reason" {
			continue
		}
		sum := sha256.Sum256(payload.Document)
		if identity.FormID != catalogDemoLostReasonID ||
			payload.ComponentID != catalogDemoLostReasonID ||
			payload.Owner != "forms" ||
			payload.Contract != "forms.catalog-definition.v1" ||
			payload.ContentSHA256 != catalogDemoLostReasonSHA ||
			hex.EncodeToString(sum[:]) != payload.ContentSHA256 {
			t.Fatalf("Sales Forms owner artifact identity differs: %+v", payload)
		}
		return catalogDemoBuildRequest(t, owner, catalogDemoRequestSpec{
			packageID: catalogDemoSalesPackageID, packageSlug: "sales-crm",
			releaseID:   catalogDemoSalesReleaseID,
			componentID: payload.ComponentID, owner: payload.Owner,
			contract: payload.Contract, contentSHA256: payload.ContentSHA256,
			document: payload.Document, idempotencyKey: "catalog-forms-sales-pg18",
		})
	}
	t.Fatal("Sales Forms owner artifact does not contain crm-lost-reason")
	return nil
}

func catalogDemoSalesCatalogRequest(t *testing.T, owner any) any {
	t.Helper()
	document := []byte(fmt.Sprintf(
		`{"component_id":"%s","contract_version":"catalog.demo-data.v1",`+
			`"demo_forms":[{"legacy_form_config_id":`+
			`"e4000000-0000-0000-0000-000000000080","rows":[`+
			`{"description":"%s","id":"%s",`+
			`"reason_name":"Prior Project Reference"}],`+
			`"slug":"crm-lost-reason"}],"excluded_form_slugs":[],`+
			`"module_slug":"sales-crm",`+
			`"source_contract":"harpy-v1.enterprise-seed.sales-crm"}`,
		catalogDemoSalesComponentID, catalogDemoLegacyB, catalogDemoCrossLegacy,
	))
	sum := sha256.Sum256(document)
	return catalogDemoBuildRequest(t, owner, catalogDemoRequestSpec{
		packageID: catalogDemoSalesPackageID, packageSlug: "sales-crm",
		releaseID:   catalogDemoSalesReleaseID,
		componentID: catalogDemoSalesComponentID, owner: "catalog",
		contract: "catalog.demo-data.v1", contentSHA256: hex.EncodeToString(sum[:]),
		document:       document,
		idempotencyKey: "catalog-demo-sales-pg18-cross-package-reference",
	})
}
