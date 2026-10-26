//go:build integration

package integrationtest

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dependency"
)

const (
	analyticsCatalogPackageID = "019f9000-0000-7000-8000-000000000001"
	analyticsCatalogReleaseID = "019f9000-0000-7000-8000-000000000002"
	analyticsCatalogActorID   = "019f9000-0000-7000-8000-000000000003"
	analyticsCatalogEndpoint  = "019f9000-0000-7000-8000-000000000010"
	analyticsCatalogDashboard = "019f9000-0000-7000-8000-000000000020"
)

type analyticsCatalogProjectionFixture struct {
	componentID, kind, slug string
	document, runtimePlan   []byte
	contentSHA              string
	semanticSHA, planSHA    [sha256.Size]byte
	auxiliaryCount          int
}

func analyticsEndpointCatalogFixture(t *testing.T) analyticsCatalogProjectionFixture {
	t.Helper()
	return analyticsEndpointCatalogFixtureFor(
		t, analyticsCatalogEndpoint, "hr-employee-directory",
	)
}

func analyticsEndpointCatalogFixtureFor(
	t *testing.T, componentID, slug string,
) analyticsCatalogProjectionFixture {
	t.Helper()
	definition := map[string]any{
		"dataSources": []any{
			map[string]any{"alias": "employees", "formSlug": "hr-employees"},
			map[string]any{"alias": "departments", "formSlug": "hr-departments"},
		},
		"fields": []any{
			map[string]any{
				"alias": "employee_id", "field": "employees.id", "source": "employees",
			},
			map[string]any{
				"alias": "department_name", "field": "departments.name",
				"source": "departments",
			},
		},
		"filters": []any{},
		"joins": []any{map[string]any{
			"joinType": "LEFT", "sourceForm": "employees",
			"sourceField": "employees.department_id", "targetForm": "departments",
			"targetField": "departments.id",
		}},
		"aggregation": map[string]any{
			"enabled": false, "aggregates": []any{},
			"groupBy": []any{}, "orderBy": []any{},
		},
	}
	plan := map[string]any{
		"contract": "analytics.catalog-runtime-plan.v1",
		"kind":     "endpoint", "status": "candidate",
		"sources": []any{"hr-employees", "hr-departments"},
		"query": map[string]any{
			"contract": "analytics.native-join-query.v1",
			"sources": []any{
				map[string]any{
					"alias": "employees", "form_slug": "hr-employees",
					"query": map[string]any{
						"select": []any{"id", "department_id"},
						"page":   map[string]any{"size": 50, "after": nil},
						"count":  "none", "include": []any{},
					},
				},
				map[string]any{
					"alias": "departments", "form_slug": "hr-departments",
					"query": map[string]any{
						"select": []any{"id", "name"},
						"page":   map[string]any{"size": 50, "after": nil},
						"count":  "none", "include": []any{},
					},
				},
			},
			"joins": []any{map[string]any{
				"kind": "left", "source": "employees", "source_field": "department_id",
				"target": "departments", "target_field": "id",
			}},
			"fields": []any{
				map[string]any{
					"alias": "employee_id", "source": "employees", "field": "id",
				},
				map[string]any{
					"alias": "department_name", "source": "departments", "field": "name",
				},
			},
			"limit": 50,
		},
		"parameters": map[string]any{}, "output": map[string]any{},
	}
	return analyticsCatalogFixture(
		t, componentID, "endpoint", slug, definition, plan, 4,
	)
}

func analyticsCatalogExpectedDependencyManifest(t *testing.T) [sha256.Size]byte {
	t.Helper()
	entries := make([]dependency.Entry, 0, 2)
	for _, fixture := range []struct {
		resourceID, versionID string
		digestByte            byte
	}{
		{"019f9000-0000-7000-8000-000000000032",
			"019f9000-0000-7000-8000-000000000033", 0x66},
		{"019f9000-0000-7000-8000-000000000035",
			"019f9000-0000-7000-8000-000000000036", 0x77},
	} {
		digest := [sha256.Size]byte{}
		for index := range digest {
			digest[index] = fixture.digestByte
		}
		entry, err := dependency.NewEntry(
			"forms", "form", analyticsCatalogUUIDBytes(t, fixture.resourceID),
			analyticsCatalogUUIDBytes(t, fixture.versionID), 1, 1, 1,
			"forms.records.revision.v1", digest,
		)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	manifest, err := dependency.NewManifest(entries, true)
	if err != nil {
		t.Fatal(err)
	}
	return manifest.SHA256()
}

func analyticsCatalogUUIDBytes(t *testing.T, value string) [16]byte {
	t.Helper()
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	if err != nil || len(decoded) != 16 {
		t.Fatalf("decode Analytics Catalog UUID %q: %v", value, err)
	}
	var result [16]byte
	copy(result[:], decoded)
	return result
}

func analyticsDashboardCatalogFixture(t *testing.T) analyticsCatalogProjectionFixture {
	t.Helper()
	definition := map[string]any{
		"dataSources": []any{map[string]any{
			"id": "employee-source", "endpointSlug": "hr-employee-directory",
		}},
		"widgets": []any{map[string]any{
			"id": "employee-table", "dataSource": "employee-source", "type": "table",
		}},
	}
	plan := map[string]any{
		"contract": "analytics.catalog-runtime-plan.v1",
		"kind":     "dashboard", "status": "candidate",
		"sources": []any{"hr-employee-directory"}, "dashboard": definition,
	}
	return analyticsCatalogFixture(
		t, analyticsCatalogDashboard, "dashboard", "hr-employee-dashboard",
		definition, plan, 2,
	)
}
