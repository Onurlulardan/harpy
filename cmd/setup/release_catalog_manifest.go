package main

import (
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const (
	localCatalogSlug = "project"
	localDemoFormID  = "e4000000-0000-0000-0000-000000000032"
	localDemoRowID   = "e5000000-0000-0000-0000-000000000032"
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

func localCatalogPayload() ([]byte, error) {
	row, err := canonicalCatalogDocument(struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{ID: localDemoRowID, Name: "Task"})
	if err != nil {
		return nil, err
	}
	return canonicalCatalogDocument(localDemoPayload{
		ContractVersion: catalogpackagev1.ContractCatalog,
		ComponentID:     localComponentID, ModuleSlug: localCatalogSlug,
		SourceContract:    catalogpackagev1.FirstPartySourceContractPrefix + localCatalogSlug,
		ExcludedFormSlugs: []string{},
		DemoForms: []localDemoForm{{
			Slug: "project-task", LegacyFormConfigID: localDemoFormID,
			Rows: []json.RawMessage{row},
		}},
	})
}

func localCatalogManifest(payloadDigest string) ([]byte, error) {
	return canonicalCatalogDocument(catalogManifestDocument{
		Contract: catalogpackagev1.ManifestContract, PackageID: localPackageID,
		Slug: localCatalogSlug, Version: "1.0.0", PlatformContract: "harpy.v2",
		Profile: "core", Dependencies: []struct{}{}, Menu: []struct{}{},
		Components: []catalogComponentWire{{
			ID: localComponentID, Owner: "catalog",
			Contract: catalogpackagev1.ContractCatalog, Kind: "DEMO_DATA",
			ContentSHA256: payloadDigest,
		}},
	})
}

func canonicalCatalogDocument(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(encoded, catalogpackagev1.MaximumPayloadBytes)
}
