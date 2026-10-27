//go:build integration

package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const catalogDemoInventoryPackageID = "e3000000-0000-0000-0000-000000000005"
const catalogDemoInventoryReleaseID = "019c3300-0000-7000-8000-000000000005"
const catalogDemoInventoryBinID = "018f0000-0000-7000-8000-000000000097"
const catalogDemoInventoryBinSHA = "5706110dcc82d991a849e57a5e62ef9bb8bea63f8bd984d46bcf94ea4e4cb5d8"
const catalogDemoAnalyticsComponentID = "018f0000-0002-7000-8000-000000004002"
const catalogDemoAnalyticsComponentSHA = "387bd54be3bed54bfebbe5a606a4b8d6e9cb79ccd7995774fb3be2e62e507eb0"

func catalogDemoInventoryFormsRequest(
	t *testing.T, root string, owner any,
) any {
	t.Helper()
	payload := catalogDemoOwnerPayload(
		t, root, "inventory", "forms", "inventory-bin",
	)
	return catalogDemoInventoryRequest(
		t, owner, payload, catalogDemoInventoryBinID,
		catalogDemoInventoryBinSHA, "catalog-forms-inventory-bin-pg18",
	)
}

func catalogDemoInventoryAnalyticsRequest(
	t *testing.T, root string, owner any,
) any {
	t.Helper()
	payload := catalogDemoOwnerPayload(
		t, root, "inventory", "analytics", "inventory-low-stock-items",
	)
	return catalogDemoInventoryRequest(
		t, owner, payload, catalogDemoAnalyticsComponentID,
		catalogDemoAnalyticsComponentSHA, "catalog-analytics-inventory-low-stock-pg18",
	)
}

func catalogDemoOwnerPayload(
	t *testing.T, root, module, owner, slug string,
) catalogDemoOwnerPayloadGroupPayload {
	t.Helper()
	path := filepath.Join(
		root, "tools", "internal", "firstpartycatalog", "testdata",
		module, "artifacts", "owner-"+owner+".json",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var group struct {
		Payloads []catalogDemoOwnerPayloadGroupPayload `json:"payloads"`
	}
	if err = json.Unmarshal(content, &group); err != nil {
		t.Fatalf("decode %s owner artifact: %v", owner, err)
	}
	for _, payload := range group.Payloads {
		var identity struct {
			Slug string `json:"slug"`
		}
		if json.Unmarshal(payload.Document, &identity) == nil &&
			identity.Slug == slug {
			return payload
		}
	}
	t.Fatalf("%s owner artifact does not contain %s", owner, slug)
	return catalogDemoOwnerPayloadGroupPayload{}
}

type catalogDemoOwnerPayloadGroupPayload struct {
	ComponentID   string          `json:"component_id"`
	ContentSHA256 string          `json:"content_sha256"`
	Contract      string          `json:"contract"`
	Document      json.RawMessage `json:"document"`
	Owner         string          `json:"owner"`
}

func catalogDemoInventoryRequest(
	t *testing.T,
	owner any,
	payload catalogDemoOwnerPayloadGroupPayload,
	componentID, contentSHA256, idempotencyKey string,
) any {
	t.Helper()
	sum := sha256.Sum256(payload.Document)
	if payload.ComponentID != componentID ||
		payload.ContentSHA256 != contentSHA256 ||
		hex.EncodeToString(sum[:]) != contentSHA256 {
		t.Fatalf("Inventory owner payload differs: %+v", payload)
	}
	return catalogDemoBuildRequest(t, owner, catalogDemoRequestSpec{
		packageID: catalogDemoInventoryPackageID, packageSlug: "inventory",
		packageVersion: "2.0.0", releaseID: catalogDemoInventoryReleaseID,
		componentID: payload.ComponentID, owner: payload.Owner,
		contract: payload.Contract, contentSHA256: payload.ContentSHA256,
		document: payload.Document, idempotencyKey: idempotencyKey,
	})
}
