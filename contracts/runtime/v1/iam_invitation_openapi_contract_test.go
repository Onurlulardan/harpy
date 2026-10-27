package runtimev1

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestIAMInvitationOpenAPIContractIsExactAndSecretFree(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	paths := readOpenAPIDocument(t, filepath.Join(directory, "iam-invitations-v1.paths.json"))
	assertLocalReferences(t, directory, paths)
	items := paths["pathItems"].(map[string]any)
	want := map[string][]string{
		"collection": {"get", "post"},
		"item":       {"get", "delete"},
		"verify":     {"get"},
		"accept":     {"post"},
	}
	if len(items) != len(want) {
		t.Fatalf("IAM invitation path item count=%d", len(items))
	}
	operations := 0
	for name, methods := range want {
		operations += len(methods)
		item := items[name].(map[string]any)
		if len(item) != len(methods) {
			t.Fatalf("IAM invitation path item %s=%v", name, item)
		}
		for _, method := range methods {
			if item[method] == nil {
				t.Fatalf("IAM invitation %s does not publish %s", name, method)
			}
		}
	}
	if operations != 6 {
		t.Fatalf("IAM invitation operation count=%d want=6", operations)
	}
	assertInvitationSecurity(t, items, "collection", "get", true)
	assertInvitationSecurity(t, items, "collection", "post", true)
	assertInvitationSecurity(t, items, "item", "get", true)
	assertInvitationSecurity(t, items, "item", "delete", true)
	assertInvitationSecurity(t, items, "verify", "get", false)
	assertInvitationSecurity(t, items, "accept", "post", false)
	acceptOperation := items["accept"].(map[string]any)["post"].(map[string]any)
	recovery := acceptOperation["responses"].(map[string]any)["409"].(map[string]any)
	if recovery["$ref"] != "iam-invitations-v1.schemas.json#/responses/AcceptanceConflict" {
		t.Fatalf("IAM invitation recovery response=%v", recovery)
	}

	schemas := readOpenAPIDocument(t, filepath.Join(directory, "iam-invitations-v1.schemas.json"))
	assertLocalReferences(t, directory, schemas)
	definitions := schemas["schemas"].(map[string]any)
	accept := definitions["AcceptRequest"].(map[string]any)
	properties := accept["properties"].(map[string]any)
	for _, name := range []string{"token", "password"} {
		property := properties[name].(map[string]any)
		if property["writeOnly"] != true || property["x-harpy-sensitive"] != true {
			t.Fatalf("IAM invitation request secret %s is not write-only: %v", name, property)
		}
	}
	parameter := schemas["parameters"].(map[string]any)["Token"].(map[string]any)
	if parameter["x-harpy-sensitive"] != true {
		t.Fatalf("IAM invitation query credential is not sensitive: %v", parameter)
	}
	recoveryResponse := schemas["responses"].(map[string]any)["AcceptanceConflict"].(map[string]any)
	codes := recoveryResponse["x-harpy-error-codes"].([]any)
	if len(codes) != 2 || codes[0] != "INVITATION_CONFLICT" ||
		codes[1] != "PROFILE_RECOVERY_REQUIRED" {
		t.Fatalf("IAM invitation recovery code=%v", recoveryResponse)
	}
	for _, name := range []string{"Invitation", "ListResponse", "VerifyResponse", "AcceptResponse", "Message"} {
		assertInvitationResponseSecretFree(t, name, definitions[name])
	}
	requiresLogin := definitions["AcceptResponse"].(map[string]any)["properties"].(map[string]any)["requires_login"].(map[string]any)
	if requiresLogin["const"] != true {
		t.Fatalf("IAM invitation acceptance must require a separate login: %v", requiresLogin)
	}
}

func assertInvitationSecurity(
	t *testing.T, items map[string]any, itemName, method string, authenticated bool,
) {
	t.Helper()
	operation := items[itemName].(map[string]any)[method].(map[string]any)
	security := operation["security"].([]any)
	if !authenticated {
		if len(security) != 0 {
			t.Fatalf("public invitation operation %s %s has security: %v", itemName, method, security)
		}
		return
	}
	if len(security) != 1 {
		t.Fatalf("protected invitation operation %s %s security=%v", itemName, method, security)
	}
	requirement := security[0].(map[string]any)
	if len(requirement) != 1 || requirement["bearerAuth"] == nil {
		t.Fatalf("protected invitation operation %s %s security=%v", itemName, method, requirement)
	}
}

func assertInvitationResponseSecretFree(t *testing.T, name string, value any) {
	t.Helper()
	wire, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(wire))
	for _, forbidden := range []string{`"token"`, `"password"`, `"session"`, `"refresh_token"`, `"access_token"`, `"secret"`, `"hmac"`, `"artifact"`} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("IAM invitation response schema %s contains %s", name, forbidden)
		}
	}
}
