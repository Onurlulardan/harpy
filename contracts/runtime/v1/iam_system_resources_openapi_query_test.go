package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestIAMSystemResourceQuerySchemasSeparatePartialAndCanonicalItems(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	records := readOpenAPIDocument(t, filepath.Join(
		directory, "iam-system-resources-v1.records.json",
	))["schemas"].(map[string]any)

	user := records["UserRecord"].(map[string]any)
	iamAssertPresent(t, user["properties"].(map[string]any), "first_name", "last_name")
	required := iamStringSet(user["required"])
	for _, name := range []string{"email", "first_name", "last_name"} {
		if !required[name] {
			t.Fatalf("user response must require published field %s", name)
		}
	}
	resource := records["ResourceRecord"].(map[string]any)
	if resource["oneOf"] != nil || len(resource["anyOf"].([]any)) != 3 {
		t.Fatalf("resource union must use three anyOf branches: %v", resource)
	}
	for _, name := range []string{
		"UserRecord", "RoleRecord", "OrganizationRecord",
		"QueryRecord", "CanonicalListRecord",
	} {
		iamAssertForbiddenNames(t, records[name].(map[string]any),
			"password", "password_hash", "email_normalized",
			"token_version", "authorization_version", "identifier_id")
	}
	canonical := records["CanonicalListRecord"].(map[string]any)
	iamAssertStrings(t, canonical["required"], "id", "revision")

	schemas := readOpenAPIDocument(t, filepath.Join(
		directory, "iam-system-resources-v1.schemas.json",
	))["schemas"].(map[string]any)
	queryData := schemas["QueryData"].(map[string]any)
	iamAssertItemReference(t, queryData, "QueryRecord")
	pagination := queryData["properties"].(map[string]any)["pagination"].(map[string]any)
	iamAssertStrings(t, pagination["required"], "has_more", "next_transport")
	nextTransport := pagination["properties"].(map[string]any)["next_transport"].(map[string]any)
	iamAssertStrings(t, nextTransport["enum"], "GET", "POST")

	canonicalData := schemas["CanonicalQueryData"].(map[string]any)
	iamAssertItemReference(t, canonicalData, "CanonicalListRecord")
}

func TestIAMSystemResourcePathsPublishCanonicalAndDirectQueryFailures(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	for _, resource := range []string{"users", "roles", "organizations"} {
		items := readOpenAPIDocument(t, filepath.Join(
			directory, "iam-"+resource+"-v1.paths.json",
		))["pathItems"].(map[string]any)
		collection := items["collection"].(map[string]any)
		iamAssertResponseReference(t, collection, "get", "200", "CanonicalQueryResult")
		iamAssertResponseCodes(t, collection, "get",
			"200", "400", "401", "403", "422", "503")
		iamAssertResponseCodes(t, collection, "post",
			"201", "400", "401", "403", "409", "413", "415", "422", "503")
		iamAssertResponseReference(t, collection, "delete", "200", "Delete")
		iamAssertResponseCodes(t, collection, "delete",
			"200", "400", "401", "403", "409", "413", "415", "503")
		iamAssertResponseCodes(t, items["query"].(map[string]any), "post",
			"200", "400", "401", "403", "413", "415", "422", "503")
		item := items["item"].(map[string]any)
		iamAssertResponseCodes(t, item, "get",
			"200", "400", "401", "403", "404", "503")
		iamAssertResponseCodes(t, item, "put",
			"200", "400", "401", "403", "404", "409", "413", "415", "422", "428", "503")
		iamAssertResponseCodes(t, item, "delete",
			"200", "400", "401", "403", "404", "409", "428", "503")
		iamAssertResponseCodes(t, items["state"].(map[string]any), "patch",
			"200", "400", "401", "403", "404", "409", "413", "415", "428", "503")
	}
}

func iamAssertItemReference(t *testing.T, data map[string]any, name string) {
	t.Helper()
	items := data["properties"].(map[string]any)["items"].(map[string]any)
	item := items["items"].(map[string]any)
	want := "iam-system-resources-v1.records.json#/schemas/" + name
	if item["$ref"] != want {
		t.Fatalf("query item ref=%v want=%s", item, want)
	}
}

func iamAssertResponseCodes(
	t *testing.T, item map[string]any, method string, expected ...string,
) {
	t.Helper()
	operation := item[method].(map[string]any)
	responses := operation["responses"].(map[string]any)
	if len(responses) != len(expected) {
		t.Fatalf("%s responses=%v want=%v", method, responses, expected)
	}
	for _, status := range expected {
		if responses[status] == nil {
			t.Fatalf("%s response is missing %s", method, status)
		}
	}
}

func iamAssertResponseReference(
	t *testing.T, item map[string]any, method, status, name string,
) {
	t.Helper()
	operation := item[method].(map[string]any)
	response := operation["responses"].(map[string]any)[status].(map[string]any)
	want := "iam-system-resources-v1.schemas.json#/responses/" + name
	if response["$ref"] != want {
		t.Fatalf("%s response %s=%v want=%s", method, status, response, want)
	}
}
