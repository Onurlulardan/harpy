package iam

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type iamOpenAPIReference struct {
	Ref string `json:"$ref"`
}

type iamOpenAPIPathDocument struct {
	PathItems map[string]map[string]json.RawMessage `json:"pathItems"`
}

type iamOpenAPIAggregate struct {
	Paths map[string]iamOpenAPIReference `json:"paths"`
}

type iamOpenAPIOperation struct {
	Deprecated bool                       `json:"deprecated"`
	Responses  map[string]json.RawMessage `json:"responses"`
}

func TestCoreOpenAPIInventoryCoversEveryComposedIAMAuthRoute(t *testing.T) {
	root := iamRepositoryRoot(t)
	fragmentPath := filepath.Join(
		root, "contracts", "http", "core-api-paths", "iam-auth.json",
	)
	fragmentBytes, err := os.ReadFile(fragmentPath)
	if err != nil {
		t.Fatal(err)
	}
	var fragment map[string]iamOpenAPIReference
	if err = json.Unmarshal(fragmentBytes, &fragment); err != nil {
		t.Fatal(err)
	}
	aggregateBytes, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "core-api-v1.openapi.yaml",
	))
	if err != nil {
		t.Fatal(err)
	}
	var aggregate iamOpenAPIAggregate
	if err = json.Unmarshal(aggregateBytes, &aggregate); err != nil {
		t.Fatal(err)
	}

	handlers := []struct {
		name   string
		routes []httpedge.Route
	}{
		{"auth", (&AuthHTTP{}).Routes()},
		{"lifecycle", (&LifecycleHTTP{}).Routes()},
		{"portal-user", (&PortalUserHTTP{}).Routes()},
	}
	retired := iamRouteSet((&LifecycleHTTP{}).retiredRoutes())
	documents := make(map[string]iamOpenAPIPathDocument)
	seen := make(map[string]string)
	documented := 0
	runtimeRoutes := 0
	for _, handler := range handlers {
		for _, route := range handler.routes {
			runtimeRoutes++
			wire := route.Method + " " + route.Pattern
			if owner, duplicate := seen[wire]; duplicate {
				t.Errorf("duplicate composed route %s in %s and %s", wire, owner, handler.name)
				continue
			}
			seen[wire] = handler.name
			reference, found := fragment[route.Pattern]
			if !found {
				t.Errorf("%s route missing from core OpenAPI fragment: %s", handler.name, wire)
				continue
			}
			filename, key, valid := iamOpenAPIReferenceTarget(reference.Ref)
			if !valid {
				t.Errorf("invalid core OpenAPI reference for %s: %q", wire, reference.Ref)
				continue
			}
			document, found := documents[filename]
			if !found {
				document = iamReadOpenAPIPathDocument(t, root, filename)
				documents[filename] = document
			}
			operations, found := document.PathItems[key]
			if !found {
				t.Errorf("path item %q missing for %s", key, wire)
				continue
			}
			operationWire, found := operations[strings.ToLower(route.Method)]
			if !found {
				t.Errorf("method missing from core OpenAPI: %s", wire)
				continue
			}
			var operation iamOpenAPIOperation
			if err = json.Unmarshal(operationWire, &operation); err != nil {
				t.Errorf("decode core OpenAPI operation for %s: %v", wire, err)
				continue
			}
			if _, isRetired := retired[wire]; isRetired {
				if !operation.Deprecated {
					t.Errorf("retired runtime route is not deprecated in OpenAPI: %s", wire)
					continue
				}
				if _, found = operation.Responses["410"]; !found {
					t.Errorf("retired runtime route has no 410 response in OpenAPI: %s", wire)
					continue
				}
			}
			aggregateReference, found := aggregate.Paths[route.Pattern]
			if !found {
				t.Errorf("runtime route missing from aggregate core OpenAPI: %s", wire)
				continue
			}
			if aggregateReference.Ref != reference.Ref {
				t.Errorf("aggregate core OpenAPI reference drift for %s: got %q want %q",
					wire, aggregateReference.Ref, reference.Ref)
				continue
			}
			documented++
		}
	}
	if documented != runtimeRoutes {
		t.Fatalf("documented IAM auth routes=%d runtime routes=%d", documented, runtimeRoutes)
	}
}

func iamOpenAPIReferenceTarget(reference string) (string, string, bool) {
	parts := strings.SplitN(reference, "#/pathItems/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func iamReadOpenAPIPathDocument(
	t *testing.T, root, filename string,
) iamOpenAPIPathDocument {
	t.Helper()
	wire, err := os.ReadFile(filepath.Join(root, "contracts", "http", filename))
	if err != nil {
		t.Fatal(err)
	}
	var document iamOpenAPIPathDocument
	if err = json.Unmarshal(wire, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func iamRouteSet(routes []httpedge.Route) map[string]struct{} {
	result := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		result[route.Method+" "+route.Pattern] = struct{}{}
	}
	return result
}

func iamRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve IAM OpenAPI test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}
