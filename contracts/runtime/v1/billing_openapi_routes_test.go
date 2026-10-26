package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type billingOperation struct {
	Method     string
	Permission string
}

func TestBillingCanonicalOpenAPIRoutesAndSecurityAreExact(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	want := canonicalBillingOperations()
	if len(want) != 34 {
		t.Fatalf("canonical operation fixture count=%d", len(want))
	}
	for key, expected := range want {
		parts := strings.SplitN(key, " ", 2)
		item := resolveCorePathItem(t, directory, paths, parts[1])
		operation, found := item[strings.ToLower(expected.Method)].(map[string]any)
		if !found {
			t.Fatalf("canonical operation %s missing: %v", key, item)
		}
		if expected.Permission == "" {
			if security, found := operation["security"].([]any); !found || len(security) != 0 {
				t.Fatalf("public operation %s security=%v", key, operation["security"])
			}
		} else {
			assertBillingBearerSecurity(t, key, operation)
			if operation["x-harpy-permission"] != expected.Permission {
				t.Fatalf("operation %s permission=%v", key, operation["x-harpy-permission"])
			}
		}
	}
}

func TestBillingV1OpenAPIMatchesExactCharacterizationRoutes(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	routes := readBillingCharacterizationRoutes(t)
	if len(routes) != 26 {
		t.Fatalf("legacy route count=%d", len(routes))
	}
	for _, route := range routes {
		path := strings.ReplaceAll(route.Path, ":id", "{id}")
		item := resolveCorePathItem(t, directory, paths, path)
		operation, found := item[strings.ToLower(route.Method)].(map[string]any)
		if !found || operation["deprecated"] != true {
			t.Fatalf("legacy operation %s %s=%v", route.Method, path, item)
		}
		if route.Auth == "public" {
			if security, found := operation["security"].([]any); !found || len(security) != 0 {
				t.Fatalf("legacy public route %s security=%v", path, operation["security"])
			}
		} else {
			assertBillingBearerSecurity(t, route.Method+" "+path, operation)
		}
	}
}

func TestBillingActiveCampaignContractPinsDeterministicEligibility(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	item := resolveCorePathItem(
		t, directory, core["paths"].(map[string]any),
		"/api/billing/v2/promotions/active",
	)
	operation := item["get"].(map[string]any)
	selection := operation["x-harpy-selection"].(map[string]any)
	order := selection["campaign_order"].([]any)
	if selection["window"] != "starts_at <= now AND (ends_at IS NULL OR now < ends_at)" ||
		len(order) != 2 || order[0] != "priority DESC" || order[1] != "campaign_version_id ASC" ||
		!strings.Contains(selection["eligibility"].(string), "not revoked") {
		t.Fatalf("active campaign selection=%v", selection)
	}
}

func resolveCorePathItem(
	t *testing.T,
	directory string,
	paths map[string]any,
	path string,
) map[string]any {
	t.Helper()
	entry, found := paths[path].(map[string]any)
	if !found {
		t.Fatalf("core path %s is missing", path)
	}
	reference, found := entry["$ref"].(string)
	if !found {
		return entry
	}
	parts := strings.SplitN(reference, "#", 2)
	document := readOpenAPIDocument(t, filepath.Join(directory, filepath.FromSlash(parts[0])))
	segments := strings.Split(strings.TrimPrefix(parts[1], "/"), "/")
	var current any = document
	for _, segment := range segments {
		current = current.(map[string]any)[segment]
	}
	return current.(map[string]any)
}

func assertBillingBearerSecurity(t *testing.T, key string, operation map[string]any) {
	t.Helper()
	security, found := operation["security"].([]any)
	if !found || len(security) != 1 {
		t.Fatalf("operation %s security=%v", key, operation["security"])
	}
	scheme := security[0].(map[string]any)
	if _, found := scheme["bearerAuth"]; !found || len(scheme) != 1 {
		t.Fatalf("operation %s security=%v", key, security)
	}
}

type characterizedBillingRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Auth   string `json:"auth"`
}

func readBillingCharacterizationRoutes(t *testing.T) []characterizedBillingRoute {
	t.Helper()
	directory := filepath.Join("..", "..", "characterization", "domains", "billing-v1", "routes")
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(files) != 6 {
		t.Fatalf("Billing V1 route files=%d err=%v", len(files), err)
	}
	var result []characterizedBillingRoute
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Routes []characterizedBillingRoute `json:"routes"`
		}
		if err := json.Unmarshal(content, &document); err != nil {
			t.Fatal(err)
		}
		result = append(result, document.Routes...)
	}
	return result
}
