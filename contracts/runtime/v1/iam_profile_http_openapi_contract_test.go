package runtimev1

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestIAMProfileHTTPOpenAPIContractIsExactAndSecretFree(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	paths := readOpenAPIDocument(t, filepath.Join(directory, "iam-profile-v1.paths.json"))
	assertLocalReferences(t, directory, paths)
	items := paths["pathItems"].(map[string]any)
	want := map[string][]string{
		"self":       {"get", "put"},
		"active":     {"get"},
		"historical": {"get"},
		"publish":    {"post"},
		"recovery":   {"post"},
		"retire":     {"post"},
	}
	if len(items) != len(want) {
		t.Fatalf("IAM profile path item count=%d want=%d", len(items), len(want))
	}
	for name, methods := range want {
		item := items[name].(map[string]any)
		if len(item) != len(methods) {
			t.Fatalf("IAM profile path item %s=%v", name, item)
		}
		for _, method := range methods {
			operation := item[method].(map[string]any)
			assertProfileBearerSecurity(t, name, method, operation)
		}
	}

	assertProfileParameterRef(t, items, "self", "put",
		"iam-profile-v1.schemas.json#/parameters/SelfIdempotencyKey")
	assertProfileParameterRef(t, items, "publish", "post",
		"iam-profile-v1.schemas.json#/parameters/IdempotencyKey")
	assertProfileParameterRef(t, items, "retire", "post",
		"iam-profile-v1.schemas.json#/parameters/IdempotencyKey")
	assertProfileParameterRef(t, items, "recovery", "post",
		"iam-profile-v1.schemas.json#/parameters/IdempotencyKey")
	assertProfileParameterRef(t, items, "recovery", "post",
		"iam-profile-v1.schemas.json#/parameters/PlanID")
	selfConflict := items["self"].(map[string]any)["put"].(map[string]any)["responses"].(map[string]any)["409"].(map[string]any)
	if selfConflict["$ref"] != "iam-profile-v1.schemas.json#/responses/ProfileWriteConflict" {
		t.Fatalf("IAM self profile recovery conflict=%v", selfConflict)
	}
	recoveryConflict := items["recovery"].(map[string]any)["post"].(map[string]any)["responses"].(map[string]any)["409"].(map[string]any)
	if recoveryConflict["$ref"] != "iam-profile-v1.schemas.json#/responses/RecoveryConflict" {
		t.Fatalf("IAM profile recovery conflict=%v", recoveryConflict)
	}

	requests := readOpenAPIDocument(t, filepath.Join(directory, "iam-profile-v1.requests.json"))
	assertLocalReferences(t, directory, requests)
	assertProfileSelfUpdateSchema(t, requests)
	assertProfileFieldEnums(t, requests)

	records := readOpenAPIDocument(t, filepath.Join(directory, "iam-profile-v1.records.json"))
	assertLocalReferences(t, directory, records)
	assertProfileRecordsSecretFree(t, records)

	schemas := readOpenAPIDocument(t, filepath.Join(directory, "iam-profile-v1.schemas.json"))
	assertLocalReferences(t, directory, schemas)
	parameters := schemas["parameters"].(map[string]any)
	if parameters["SelfIdempotencyKey"].(map[string]any)["required"] != false ||
		parameters["IdempotencyKey"].(map[string]any)["required"] != true {
		t.Fatal("IAM profile idempotency requirements are not exact")
	}
	assertProfileErrorCodes(t, schemas)
	profileConflict := schemas["responses"].(map[string]any)["ProfileWriteConflict"].(map[string]any)
	codes := profileConflict["x-harpy-error-codes"].([]any)
	if len(codes) != 2 || codes[0] != "RESOURCE_CONFLICT" ||
		codes[1] != "PROFILE_RECOVERY_REQUIRED" {
		t.Fatalf("IAM profile write conflict codes=%v", codes)
	}
	recoveryCodes := schemas["responses"].(map[string]any)["RecoveryConflict"].(map[string]any)["x-harpy-error-codes"].([]any)
	if len(recoveryCodes) != 2 || recoveryCodes[0] != "RESOURCE_CONFLICT" ||
		recoveryCodes[1] != "PROFILE_RECOVERY_BUDGET_EXCEEDED" {
		t.Fatalf("IAM recovery conflict codes=%v", recoveryCodes)
	}
}

func assertProfileBearerSecurity(
	t *testing.T, item, method string, operation map[string]any,
) {
	t.Helper()
	security := operation["security"].([]any)
	if len(security) != 1 {
		t.Fatalf("IAM profile %s %s security=%v", item, method, security)
	}
	requirement := security[0].(map[string]any)
	if len(requirement) != 1 || requirement["bearerAuth"] == nil {
		t.Fatalf("IAM profile %s %s security=%v", item, method, requirement)
	}
}

func assertProfileParameterRef(
	t *testing.T, items map[string]any, item, method, reference string,
) {
	t.Helper()
	operation := items[item].(map[string]any)[method].(map[string]any)
	for _, raw := range operation["parameters"].([]any) {
		if raw.(map[string]any)["$ref"] == reference {
			return
		}
	}
	t.Fatalf("IAM profile %s %s lacks parameter %s", item, method, reference)
}

func assertProfileSelfUpdateSchema(t *testing.T, document map[string]any) {
	t.Helper()
	schema := document["schemas"].(map[string]any)["SelfUpdate"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if schema["additionalProperties"] != false || schema["minProperties"] != float64(1) ||
		schema["maxProperties"] != float64(4) || len(properties) != 4 {
		t.Fatalf("IAM self profile update schema is not exact: %v", schema)
	}
	for _, name := range []string{"first_name", "last_name", "phone", "avatar"} {
		if properties[name] == nil {
			t.Fatalf("IAM self profile update lacks %s", name)
		}
	}
}

func assertProfileFieldEnums(t *testing.T, document map[string]any) {
	t.Helper()
	field := document["schemas"].(map[string]any)["Field"].(map[string]any)
	properties := field["properties"].(map[string]any)
	want := map[string]int{
		"type": 9, "cardinality": 2, "classification": 4,
		"normalization": 3, "state": 2,
	}
	for name, count := range want {
		values := properties[name].(map[string]any)["enum"].([]any)
		if len(values) != count {
			t.Fatalf("IAM profile field %s enum=%v", name, values)
		}
	}
}

func assertProfileRecordsSecretFree(t *testing.T, document map[string]any) {
	t.Helper()
	definitions := document["schemas"].(map[string]any)
	for _, name := range []string{"SelfProfile", "Schema", "Mutation", "Recovery"} {
		wire, err := json.Marshal(definitions[name])
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(wire))
		for _, forbidden := range []string{
			`"password"`, `"password_hash"`, `"email_normalized"`,
			`"token"`, `"session"`, `"secret"`,
		} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("IAM profile response schema %s contains %s", name, forbidden)
			}
		}
	}
}

func assertProfileErrorCodes(t *testing.T, document map[string]any) {
	t.Helper()
	responses := document["responses"].(map[string]any)
	want := map[string]string{
		"RequestInvalid": "REQUEST_INVALID", "AuthenticationFailed": "AUTHENTICATION_FAILED",
		"Forbidden": "AUTHORIZATION_DENIED", "FieldForbidden": "PROFILE_FIELD_FORBIDDEN",
		"NotFound": "RESOURCE_NOT_FOUND", "Conflict": "RESOURCE_CONFLICT",
		"BodyTooLarge": "REQUEST_BODY_TOO_LARGE", "MediaTypeUnsupported": "MEDIA_TYPE_UNSUPPORTED",
		"Unavailable": "IAM_UNAVAILABLE",
	}
	for name, code := range want {
		response := responses[name].(map[string]any)
		if response["x-harpy-error-code"] != code {
			t.Fatalf("IAM profile error %s=%v", name, response)
		}
	}
}
