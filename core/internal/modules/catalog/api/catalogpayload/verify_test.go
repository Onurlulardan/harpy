package catalogpayload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

func TestVerifyAcceptsPayloadDerivedFirstPartyDemoComposition(t *testing.T) {
	componentID := "018f0000-0004-7000-8000-000000000009"
	row := canonicalCatalogJSON(t, map[string]any{
		"id": "e5000000-0000-0000-0000-000000000009", "name": "Task",
	})
	payload := Payload{
		ContractVersion: catalogpackagev1.ContractCatalog,
		ComponentID:     componentID, ModuleSlug: "project",
		SourceContract:    "harpy-v1.enterprise-seed.project",
		ExcludedFormSlugs: []string{"project-obsolete"},
		DemoForms: []DemoForm{{
			Slug:               "project-task",
			LegacyFormConfigID: "e4000000-0000-0000-0000-000000000009",
			Rows:               []json.RawMessage{row},
		}},
	}
	document := canonicalCatalogJSON(t, payload)
	digest := sha256.Sum256(document)
	verified, err := Verify(document, componentID, hex.EncodeToString(digest[:]))
	if err != nil || len(verified.Value().DemoForms) != 1 {
		t.Fatalf("verify project demo: value=%#v err=%v", verified.Value(), err)
	}
	payload.DemoForms[0].LegacyFormConfigID =
		"00000000-0000-0000-0000-000000000000"
	nilIdentity := canonicalCatalogJSON(t, payload)
	nilIdentityDigest := sha256.Sum256(nilIdentity)
	if _, err = Verify(
		nilIdentity, componentID, hex.EncodeToString(nilIdentityDigest[:]),
	); err == nil {
		t.Fatal("catalog accepted a nil legacy form identity")
	}
	payload.DemoForms[0].LegacyFormConfigID =
		"e4000000-0000-0000-0000-000000000009"
	payload.DemoForms[0].Rows = []json.RawMessage{canonicalCatalogJSON(
		t, map[string]any{
			"id": "00000000-0000-0000-0000-000000000000", "name": "nil",
		},
	)}
	nilRow := canonicalCatalogJSON(t, payload)
	nilRowDigest := sha256.Sum256(nilRow)
	if _, err = Verify(
		nilRow, componentID, hex.EncodeToString(nilRowDigest[:]),
	); err == nil {
		t.Fatal("catalog accepted a nil legacy row identity")
	}
	payload.DemoForms[0].Rows = []json.RawMessage{row}
	payload.ExcludedFormSlugs = []string{"project-task"}
	mismatched := canonicalCatalogJSON(t, payload)
	mismatchedDigest := sha256.Sum256(mismatched)
	if _, err = Verify(
		mismatched, componentID, hex.EncodeToString(mismatchedDigest[:]),
	); err == nil {
		t.Fatal("catalog accepted an excluded form that is also materialized")
	}
}

func canonicalCatalogJSON(t *testing.T, value any) []byte {
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
