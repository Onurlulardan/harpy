package api

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type layoutHTTPPathsFragment struct {
	OpenAPI string                                `json:"openapi"`
	Paths   map[string]map[string]json.RawMessage `json:"paths"`
}

type layoutHTTPContractOperation struct {
	Permission string                     `json:"x-harpy-permission"`
	Authority  string                     `json:"x-harpy-authority"`
	Security   []map[string][]string      `json:"security"`
	Responses  map[string]json.RawMessage `json:"responses"`
}

type layoutHTTPReference struct {
	Reference string `json:"$ref"`
}

func TestHTTPContractFragmentsMatchRoutesAndAuthority(t *testing.T) {
	pathsDocument, err := os.ReadFile(
		"../../../../contracts/http/system-resource-layout-v2.paths.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	var fragment layoutHTTPPathsFragment
	if err = json.Unmarshal(pathsDocument, &fragment); err != nil {
		t.Fatal(err)
	}
	if fragment.OpenAPI != "3.1.1" {
		t.Fatalf("OpenAPI version = %q", fragment.OpenAPI)
	}
	schemasDocument, err := os.ReadFile(
		"../../../../contracts/http/system-resource-layout-v2.schemas.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	var schemas struct {
		Definitions map[string]json.RawMessage `json:"$defs"`
	}
	if json.Unmarshal(schemasDocument, &schemas) != nil ||
		schemas.Definitions["TenantIDParameter"] == nil ||
		schemas.Definitions["TenantSlugParameter"] == nil ||
		schemas.Definitions["ErrorResponse"] == nil {
		t.Fatal("system-resource layout schemas fragment is invalid")
	}
	want := map[string][2]string{
		"GET " + layoutHTTPCollectionPath:   {"iam.{resource}.read", "system_resource_layout.list"},
		"PUT " + layoutHTTPRevisionPath:     {"iam.{resource}.update", "system_resource_layout.create"},
		"GET " + layoutHTTPRevisionPath:     {"iam.{resource}.read", "system_resource_layout.get"},
		"PATCH " + layoutHTTPRevisionPath:   {"iam.{resource}.update", "system_resource_layout.update"},
		"DELETE " + layoutHTTPRevisionPath:  {"iam.{resource}.update", "system_resource_layout.delete"},
		"PUT " + layoutHTTPPublicationPath:  {"iam.{resource}.update", "system_resource_layout.publish"},
		"GET " + layoutHTTPPresentationPath: {"iam.{resource}.read", "system_resource_layout.resolve"},
	}
	for _, route := range new(systemResourceLayoutHTTPHandler).Routes() {
		methods, found := fragment.Paths[route.Pattern]
		if !found {
			t.Fatalf("contract path missing: %s", route.Pattern)
		}
		document, found := methods[strings.ToLower(route.Method)]
		if !found {
			t.Fatalf("contract operation missing: %s %s", route.Method, route.Pattern)
		}
		var operation layoutHTTPContractOperation
		if json.Unmarshal(document, &operation) != nil {
			t.Fatal("contract operation is invalid")
		}
		expected := want[route.Method+" "+route.Pattern]
		if operation.Permission != expected[0] || operation.Authority != expected[1] {
			t.Fatalf("contract authority %s %s = %+v", route.Method, route.Pattern, operation)
		}
		if len(operation.Security) != 1 ||
			operation.Security[0]["bearerAuth"] == nil ||
			len(operation.Security[0]["bearerAuth"]) != 0 {
			t.Fatalf("contract bearer security missing: %s %s", route.Method, route.Pattern)
		}
		var failure layoutHTTPReference
		if json.Unmarshal(operation.Responses["default"], &failure) != nil ||
			failure.Reference != "system-resource-layout-v2.schemas.json#/$defs/ErrorResponse" {
			t.Fatalf("contract default error missing: %s %s", route.Method, route.Pattern)
		}
	}
	for path, item := range fragment.Paths {
		var parameters []layoutHTTPReference
		if json.Unmarshal(item["parameters"], &parameters) != nil ||
			!layoutHTTPHasReference(parameters,
				"system-resource-layout-v2.schemas.json#/$defs/TenantIDParameter") ||
			!layoutHTTPHasReference(parameters,
				"system-resource-layout-v2.schemas.json#/$defs/TenantSlugParameter") {
			t.Fatalf("tenant headers missing from path %s", path)
		}
	}
}

func layoutHTTPHasReference(parameters []layoutHTTPReference, expected string) bool {
	for _, parameter := range parameters {
		if parameter.Reference == expected {
			return true
		}
	}
	return false
}
