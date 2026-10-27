package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestIAMAuthOpenAPIContractIsExactAndResolvable(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	paths := readOpenAPIDocument(t, filepath.Join(directory, "iam-auth-v1.paths.json"))
	assertLocalReferences(t, directory, paths)
	pathItems := paths["pathItems"].(map[string]any)
	wantMethods := map[string][]string{
		"operatorLogin":            {"post"},
		"portalLogin":              {"post"},
		"refresh":                  {"post"},
		"logout":                   {"post"},
		"csrf":                     {"get", "post"},
		"currentSession":           {"get"},
		"reauthenticationStart":    {"post"},
		"reauthenticationComplete": {"post"},
	}
	if len(pathItems) != len(wantMethods) {
		t.Fatalf("IAM auth path item count=%d", len(pathItems))
	}
	for name, methods := range wantMethods {
		item, found := pathItems[name].(map[string]any)
		if !found || len(item) != len(methods) {
			t.Fatalf("IAM auth path item %s is not exact: %v", name, item)
		}
		for _, method := range methods {
			if item[method] == nil {
				t.Fatalf("IAM auth path item %s misses %s: %v", name, method, item)
			}
		}
	}
	schemas := readOpenAPIDocument(t, filepath.Join(directory, "iam-auth-v1.schemas.json"))
	assertLocalReferences(t, directory, schemas)
	parameters := schemas["parameters"].(map[string]any)
	assertIAMAuthParameterName(t, parameters, "RefreshCookie", "__Secure-harpy-refresh")
	assertIAMAuthParameterName(t, parameters, "CSRFCookie", "__Host-harpy-csrf")
	assertIAMAuthParameterName(t, parameters, "CSRFHeader", "X-CSRF-Token")
	security := schemas["securitySchemes"].(map[string]any)["bearerAuth"].(map[string]any)
	if security["type"] != "http" || security["scheme"] != "bearer" {
		t.Fatalf("unexpected IAM bearer contract: %v", security)
	}
	assertCurrentSessionCapabilityContract(t, schemas)
	assertReauthenticationContract(t, schemas)
	responses := schemas["responses"].(map[string]any)
	assertIAMAuthErrorCode(t, responses, "AuthenticationFailed", "AUTHENTICATION_FAILED")
	assertIAMAuthErrorCode(t, responses, "CSRFInvalid", "AUTH_CSRF_INVALID")
	assertIAMAuthErrorCode(t, responses, "AuthenticationUnavailable", "AUTHENTICATION_UNAVAILABLE")
}

func assertReauthenticationContract(t *testing.T, document map[string]any) {
	t.Helper()
	schemas := document["schemas"].(map[string]any)
	start := schemas["ReauthenticationStartRequest"].(map[string]any)
	complete := schemas["ReauthenticationCompleteRequest"].(map[string]any)
	challenge := schemas["ReauthenticationChallenge"].(map[string]any)
	for name, schema := range map[string]map[string]any{
		"start": start, "complete": complete, "challenge": challenge,
	} {
		if schema["additionalProperties"] != false {
			t.Fatalf("reauthentication %s schema is not exact", name)
		}
	}
	properties := challenge["properties"].(map[string]any)
	policy := properties["action_policy_version"].(map[string]any)
	if policy["const"] != float64(1) {
		t.Fatalf("reauthentication policy version=%v", policy)
	}
}

func assertCurrentSessionCapabilityContract(t *testing.T, document map[string]any) {
	t.Helper()
	schema := document["schemas"].(map[string]any)["CurrentSession"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	for _, required := range []string{
		"deployment_mode", "platform_capabilities", "tenant_capabilities",
		"password_credential_revision",
	} {
		if properties[required] == nil {
			t.Fatalf("current session capability field %q is missing", required)
		}
	}
	required := schema["required"].([]any)
	if !containsJSONText(required, "password_credential_revision") {
		t.Fatal("password credential revision is not required")
	}
	if properties["permissions"] != nil {
		t.Fatal("current session still publishes merged permissions")
	}
}

func containsJSONText(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertIAMAuthParameterName(
	t *testing.T,
	parameters map[string]any,
	key string,
	want string,
) {
	t.Helper()
	parameter, found := parameters[key].(map[string]any)
	if !found || parameter["name"] != want || parameter["required"] != true {
		t.Fatalf("IAM auth parameter %s=%v", key, parameter)
	}
}

func assertIAMAuthErrorCode(
	t *testing.T,
	responses map[string]any,
	key string,
	want string,
) {
	t.Helper()
	response, found := responses[key].(map[string]any)
	if !found || response["x-harpy-error-code"] != want {
		t.Fatalf("IAM auth error %s=%v", key, response)
	}
}
