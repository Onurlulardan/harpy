package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestIAMSystemResourceMutationSchemasMatchCanonicalEdge(t *testing.T) {
	document := readOpenAPIDocument(t, filepath.Join(
		"..", "..", "http", "iam-system-resources-v1.requests.json",
	))
	schemas := document["schemas"].(map[string]any)

	userCreate := schemas["UserCreate"].(map[string]any)
	iamAssertStrings(t, userCreate["required"], "email", "password", "first_name", "last_name")
	userCreateProperties := userCreate["properties"].(map[string]any)
	iamAssertPresent(t, userCreateProperties, "first_name", "last_name")
	password := userCreateProperties["password"].(map[string]any)
	if password["minLength"] != float64(15) ||
		password["maxLength"] != float64(128) {
		t.Fatalf("user password scalar limits=%v..%v",
			password["minLength"], password["maxLength"])
	}
	iamAssertForbiddenNames(t, userCreate,
		"password_hash", "email_normalized", "token_version",
		"authorization_version", "identifier_id", "revision")

	userUpdate := schemas["UserUpdate"].(map[string]any)
	userUpdateProperties := userUpdate["properties"].(map[string]any)
	iamAssertPresent(t, userUpdateProperties, "first_name", "last_name")
	organizationAlias := userUpdateProperties["organization_id"].(map[string]any)
	if organizationAlias["format"] != "uuid" || organizationAlias["deprecated"] != true {
		t.Fatalf("user organization compatibility alias=%v", organizationAlias)
	}
	iamAssertForbiddenNames(t, userUpdate,
		"email", "password", "status", "revision", "created_at", "updated_at")

	roleCreate := schemas["RoleCreate"].(map[string]any)
	iamAssertStrings(t, roleCreate["required"], "name")
	scopeKind := roleCreate["properties"].(map[string]any)["scope_kind"].(map[string]any)
	iamAssertStrings(t, scopeKind["enum"], "tenant", "organization")
	roleUpdate := schemas["RoleUpdate"].(map[string]any)
	iamAssertForbiddenNames(t, roleUpdate,
		"scope_kind", "scope_id", "organization_id", "is_protected", "status")

	for _, name := range []string{
		"RoleCreate", "RoleUpdate", "OrganizationCreate", "OrganizationUpdate",
	} {
		schema := schemas[name].(map[string]any)
		if schema["additionalProperties"] != true {
			t.Fatalf("%s must admit catalog-published profile fields", name)
		}
		iamAssertForbiddenNames(t, schema,
			"password", "password_hash", "email_normalized",
			"token_version", "authorization_version", "identifier_id")
	}
}

func iamAssertForbiddenNames(t *testing.T, schema map[string]any, names ...string) {
	t.Helper()
	propertyNames := schema["propertyNames"].(map[string]any)
	not := propertyNames["not"].(map[string]any)
	actual := iamStringSet(not["enum"])
	for _, name := range names {
		if !actual[name] {
			t.Fatalf("schema does not machine-forbid property %q", name)
		}
	}
}

func iamAssertStrings(t *testing.T, value any, expected ...string) {
	t.Helper()
	actual := iamStringSet(value)
	if len(actual) != len(expected) {
		t.Fatalf("string set=%v want=%v", actual, expected)
	}
	for _, item := range expected {
		if !actual[item] {
			t.Fatalf("string set=%v is missing %q", actual, item)
		}
	}
}

func iamStringSet(value any) map[string]bool {
	result := map[string]bool{}
	for _, item := range value.([]any) {
		result[item.(string)] = true
	}
	return result
}

func iamAssertPresent(t *testing.T, values map[string]any, names ...string) {
	t.Helper()
	for _, name := range names {
		property, ok := values[name].(map[string]any)
		if !ok || property["type"] != "string" {
			t.Fatalf("published string property %q is missing", name)
		}
	}
}
