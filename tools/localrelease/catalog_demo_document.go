package main

import (
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const (
	localDemoFormID = "e4000000-0000-0000-0000-000000000032"
	localDemoRowID  = "e5000000-0000-0000-0000-000000000032"
)

type catalogManifestDocument struct {
	Contract         string                 `json:"contract"`
	PackageID        string                 `json:"package_id"`
	Slug             string                 `json:"slug"`
	Version          string                 `json:"version"`
	PlatformContract string                 `json:"platform_contract"`
	Profile          string                 `json:"profile"`
	Dependencies     []struct{}             `json:"dependencies"`
	Components       []catalogComponentWire `json:"components"`
	Menu             []struct{}             `json:"menu"`
}

type catalogComponentWire struct {
	ID            string `json:"ID"`
	Owner         string `json:"Owner"`
	Contract      string `json:"Contract"`
	Kind          string `json:"kind"`
	ContentSHA256 string `json:"content_sha256"`
}

type localDemoPayload struct {
	ContractVersion   string          `json:"contract_version"`
	ComponentID       string          `json:"component_id"`
	ModuleSlug        string          `json:"module_slug"`
	SourceContract    string          `json:"source_contract"`
	ExcludedFormSlugs []string        `json:"excluded_form_slugs"`
	DemoForms         []localDemoForm `json:"demo_forms"`
}

type localDemoForm struct {
	Slug               string            `json:"slug"`
	LegacyFormConfigID string            `json:"legacy_form_config_id"`
	Rows               []json.RawMessage `json:"rows"`
}

func localDemoDocuments(payloadDigest string) (manifest, payload []byte, err error) {
	row, err := canonicalCatalogDocument(struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{ID: localDemoRowID, Name: "Task"})
	if err != nil {
		return nil, nil, err
	}
	payload, err = canonicalCatalogDocument(localDemoPayload{
		ContractVersion: catalogpackagev1.ContractCatalog,
		ComponentID:     localCatalogComponentID, ModuleSlug: "project",
		SourceContract:    catalogpackagev1.FirstPartySourceContractPrefix + "project",
		ExcludedFormSlugs: []string{},
		DemoForms: []localDemoForm{{
			Slug: "project-task", LegacyFormConfigID: localDemoFormID,
			Rows: []json.RawMessage{row},
		}},
	})
	if err != nil {
		return nil, nil, err
	}
	manifest, err = canonicalCatalogDocument(catalogManifestDocument{
		Contract: catalogpackagev1.ManifestContract, PackageID: localCatalogPackageID,
		Slug: "project", Version: "1.0.0", PlatformContract: "harpy.v2",
		Profile: "core", Dependencies: []struct{}{}, Menu: []struct{}{},
		Components: []catalogComponentWire{{
			ID: localCatalogComponentID, Owner: "catalog",
			Contract: catalogpackagev1.ContractCatalog, Kind: "DEMO_DATA",
			ContentSHA256: payloadDigest,
		}},
	})
	return manifest, payload, err
}

func canonicalCatalogDocument(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(encoded, catalogpackagev1.MaximumPayloadBytes)
}
