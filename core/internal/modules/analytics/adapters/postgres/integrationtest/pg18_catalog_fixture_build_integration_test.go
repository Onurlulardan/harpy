//go:build integration

package integrationtest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const analyticsCatalogSourceContract = "harpy-v1.enterprise-seed.hr"

func analyticsCatalogFixture(
	t *testing.T, componentID, kind, slug string, definition, rawPlan any,
	auxiliaryCount int,
) analyticsCatalogProjectionFixture {
	t.Helper()
	name := slug
	description := "PostgreSQL operational projection fixture"
	semantic := analyticsCatalogLegacyDigest(t, slug, name, description, definition)
	original := map[string]any{
		"contract_version": "analytics.catalog-definition.v1",
		"component_id":     componentID, "analytics_id": componentID,
		"kind": kind, "slug": slug, "name": name, "description": description,
		"legacy_hash":     hex.EncodeToString(semantic[:]),
		"source_contract": analyticsCatalogSourceContract, "definition": definition,
	}
	document := analyticsCatalogJSON(t, original)
	content := sha256.Sum256(document)
	plan := analyticsCatalogMap(t, rawPlan)
	plan["component_id"] = componentID
	plan["source_contract"] = analyticsCatalogSourceContract
	plan["document_sha256"] = hex.EncodeToString(content[:])
	runtimePlan := analyticsCatalogJSON(t, plan)
	return analyticsCatalogProjectionFixture{
		componentID: componentID, kind: kind, slug: slug,
		document: document, runtimePlan: runtimePlan,
		contentSHA: hex.EncodeToString(content[:]), semanticSHA: semantic,
		planSHA: sha256.Sum256(runtimePlan), auxiliaryCount: auxiliaryCount,
	}
}

func analyticsCatalogVersionFixture(
	t *testing.T, source analyticsCatalogProjectionFixture, marker string,
	mutate func(map[string]any),
) analyticsCatalogProjectionFixture {
	t.Helper()
	original := analyticsCatalogMap(t, json.RawMessage(source.document))
	definition, ok := original["definition"].(map[string]any)
	if !ok {
		t.Fatal("Analytics Catalog fixture definition is invalid")
	}
	if mutate != nil {
		mutate(definition)
	}
	description := original["description"].(string) + " " + marker
	original["description"] = description
	semantic := analyticsCatalogLegacyDigest(
		t, source.slug, original["name"].(string), description, definition,
	)
	original["legacy_hash"] = hex.EncodeToString(semantic[:])
	document := analyticsCatalogJSON(t, original)
	content := sha256.Sum256(document)
	plan := analyticsCatalogMap(t, json.RawMessage(source.runtimePlan))
	plan["document_sha256"] = hex.EncodeToString(content[:])
	if source.kind == "dashboard" {
		plan["dashboard"] = definition
	}
	runtimePlan := analyticsCatalogJSON(t, plan)
	source.document = document
	source.runtimePlan = runtimePlan
	source.contentSHA = hex.EncodeToString(content[:])
	source.semanticSHA = semantic
	source.planSHA = sha256.Sum256(runtimePlan)
	source.auxiliaryCount = analyticsCatalogAuxiliaryCount(t, source.kind, definition)
	return source
}

func analyticsCatalogRuntimePlanFixture(
	t *testing.T, source analyticsCatalogProjectionFixture,
	mutate func(map[string]any),
) analyticsCatalogProjectionFixture {
	t.Helper()
	plan := analyticsCatalogMap(t, json.RawMessage(source.runtimePlan))
	mutate(plan)
	source.runtimePlan = analyticsCatalogJSON(t, plan)
	source.planSHA = sha256.Sum256(source.runtimePlan)
	return source
}

func analyticsCatalogLegacyDigest(
	t *testing.T, slug, name, description string, definition any,
) [sha256.Size]byte {
	t.Helper()
	return sha256.Sum256(analyticsCatalogJSON(t, map[string]any{
		"slug": slug, "name": name, "description": description,
		"definition": definition,
	}))
}

func analyticsCatalogAuxiliaryCount(
	t *testing.T, kind string, definition map[string]any,
) int {
	t.Helper()
	sources, sourceOK := definition["dataSources"].([]any)
	key := "fields"
	if kind == "dashboard" {
		key = "widgets"
	}
	items, itemOK := definition[key].([]any)
	if !sourceOK || !itemOK {
		t.Fatal("Analytics Catalog fixture composition is invalid")
	}
	return len(sources) + len(items)
}

func analyticsCatalogMap(t *testing.T, value any) map[string]any {
	t.Helper()
	encoded := analyticsCatalogJSON(t, value)
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func analyticsCatalogJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
