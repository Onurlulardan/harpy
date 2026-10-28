package main

import (
	"encoding/json"
	"testing"
	"time"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestCatalogSetSatisfiesReleaseFinalizer(t *testing.T) {
	keys, err := newReleaseKeys(newReleaseClock(time.Date(2026, 7, 22, 8, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	document, err := buildCatalogSet(keys)
	if err != nil {
		t.Fatal(err)
	}
	var set catalogpackagev1.ReleaseSet
	if err = attestation.DecodeStrict(document, &set); err != nil {
		t.Fatal(err)
	}
	if len(set.Packages) != 1 || len(set.Packages[0].OwnerPayloads) != 1 {
		t.Fatal("catalog set did not carry one sealed package")
	}
	sealed := set.Packages[0]
	var manifest struct {
		Contract   string `json:"contract"`
		Slug       string `json:"slug"`
		Components []struct {
			ID       string `json:"ID"`
			Owner    string `json:"Owner"`
			Contract string `json:"Contract"`
			Kind     string `json:"kind"`
			Hash     string `json:"content_sha256"`
		} `json:"components"`
	}
	if err = json.Unmarshal(sealed.Manifest, &manifest); err != nil {
		t.Fatal(err)
	}
	payload := sealed.OwnerPayloads[0]
	var demo struct {
		ContractVersion string `json:"contract_version"`
		ComponentID     string `json:"component_id"`
		ModuleSlug      string `json:"module_slug"`
		SourceContract  string `json:"source_contract"`
	}
	if err = json.Unmarshal(payload.Document, &demo); err != nil {
		t.Fatal(err)
	}
	if manifest.Contract != catalogpackagev1.ManifestContract ||
		manifest.Slug != localCatalogSlug || len(manifest.Components) != 1 ||
		manifest.Components[0].Owner != "catalog" ||
		manifest.Components[0].Kind != "DEMO_DATA" ||
		manifest.Components[0].Contract != catalogpackagev1.ContractCatalog ||
		manifest.Components[0].ID != payload.ComponentID ||
		manifest.Components[0].Hash != payload.ContentSHA256 ||
		demo.ContractVersion != catalogpackagev1.ContractCatalog ||
		demo.ComponentID != payload.ComponentID ||
		demo.ModuleSlug != manifest.Slug ||
		demo.SourceContract != catalogpackagev1.FirstPartySourceContractPrefix+manifest.Slug {
		t.Fatal("catalog manifest does not bind the demo payload")
	}
}
