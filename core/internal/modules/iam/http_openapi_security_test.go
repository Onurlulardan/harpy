package iam

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type iamOpenAPISecurityOperation struct {
	Security   []map[string][]string `json:"security"`
	Parameters []iamOpenAPIReference `json:"parameters"`
	Permission json.RawMessage       `json:"x-harpy-permission"`
}

func TestIAMLifecycleOpenAPISecurityMatchesHandlerContract(t *testing.T) {
	root := iamRepositoryRoot(t)
	wire, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "core-api-paths", "iam-auth.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	var fragment map[string]iamOpenAPIReference
	if err = json.Unmarshal(wire, &fragment); err != nil {
		t.Fatal(err)
	}
	documents := make(map[string]iamOpenAPIPathDocument)
	routes := append(append(
		(&AuthHTTP{}).Routes(), (&LifecycleHTTP{}).Routes()...,
	), (&PortalUserHTTP{}).Routes()...)
	cas := iamOpenAPICASRoutes()
	for _, route := range routes {
		reference, found := fragment[route.Pattern]
		if !found {
			continue
		}
		filename, key, valid := iamOpenAPIReferenceTarget(reference.Ref)
		if !valid || (!strings.HasPrefix(filename, "iam-lifecycle-") &&
			filename != "iam-portal-users-v1.paths.json") {
			continue
		}
		document, found := documents[filename]
		if !found {
			document = iamReadOpenAPIPathDocument(t, root, filename)
			documents[filename] = document
		}
		raw := document.PathItems[key][strings.ToLower(route.Method)]
		var operation iamOpenAPISecurityOperation
		if err = json.Unmarshal(raw, &operation); err != nil {
			t.Fatalf("decode %s %s security contract: %v", route.Method, route.Pattern, err)
		}
		name := route.Method + " " + route.Pattern
		authenticated := filename != "iam-lifecycle-public-v1.paths.json" &&
			filename != "iam-lifecycle-retired-v1.paths.json"
		if !authenticated {
			if len(operation.Security) != 0 || len(operation.Permission) != 0 {
				t.Errorf("public or retired route declares authority: %s", name)
			}
			continue
		}
		if !iamBearerSecurity(operation.Security) {
			t.Errorf("authenticated route lacks exact bearer security: %s", name)
		}
		parameters := iamOpenAPIParameterSet(operation.Parameters)
		if _, found = parameters["TenantID"]; !found {
			t.Errorf("authenticated route lacks optional tenant fence: %s", name)
		}
		if len(operation.Permission) == 0 || string(operation.Permission) == "null" {
			t.Errorf("authenticated route lacks handler-derived permission: %s", name)
		}
		mutation := route.Method != "GET"
		_, hasIdempotency := parameters["IdempotencyKey"]
		if mutation != hasIdempotency {
			t.Errorf("idempotency contract drift for %s: present=%t", name, hasIdempotency)
		}
		_, wantsCAS := cas[name]
		_, hasCAS := parameters["IfMatch"]
		if wantsCAS != hasCAS {
			t.Errorf("CAS contract drift for %s: present=%t", name, hasCAS)
		}
	}
}

func TestInitialPasswordOpenAPIDeclaresEveryBrowserProof(t *testing.T) {
	root := iamRepositoryRoot(t)
	document := iamReadOpenAPIPathDocument(t, root, "iam-lifecycle-public-v1.paths.json")
	raw := document.PathItems["initialPassword"]["post"]
	var operation iamOpenAPISecurityOperation
	if err := json.Unmarshal(raw, &operation); err != nil {
		t.Fatal(err)
	}
	parameters := iamOpenAPIParameterSet(operation.Parameters)
	for _, required := range []string{
		"Origin", "InitialCSRF", "InitialChangeCookie", "InitialCSRFCookie",
	} {
		if _, found := parameters[required]; !found {
			t.Errorf("initial password browser proof missing: %s", required)
		}
	}
}

func iamBearerSecurity(security []map[string][]string) bool {
	if len(security) != 1 {
		return false
	}
	scopes, found := security[0]["bearerAuth"]
	return found && len(scopes) == 0
}

func iamOpenAPIParameterSet(parameters []iamOpenAPIReference) map[string]struct{} {
	result := make(map[string]struct{}, len(parameters))
	for _, parameter := range parameters {
		index := strings.LastIndex(parameter.Ref, "/")
		if index >= 0 && index+1 < len(parameter.Ref) {
			result[parameter.Ref[index+1:]] = struct{}{}
		}
	}
	return result
}

func iamOpenAPICASRoutes() map[string]struct{} {
	values := []string{
		"POST /api/iam/identifiers/{id}/tombstone",
		"POST /api/iam/identifiers/{id}/retire",
		"POST /api/iam/password/change",
		"POST /api/iam/portal-setups/{binding_id}/decision",
		"POST /api/iam/authenticators/webauthn/enrollments/{authenticator_id}/complete",
		"POST /api/iam/authenticators/totp/enrollments/{authenticator_id}/complete",
		"POST /api/iam/authenticators/{id}/suspend",
		"POST /api/iam/authenticators/{id}/retire",
		"DELETE /api/iam/sessions/{id}",
		"PATCH /api/auth/portal-users/{id}/status",
		"POST /api/auth/portal-users/{id}/console-access",
		"DELETE /api/auth/portal-users/{id}/console-access",
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
