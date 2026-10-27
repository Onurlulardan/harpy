package runtimev1

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestCursorKeyringSchemaAcceptsExactWire(t *testing.T) {
	schema := compileSchema(t, "rest-query-cursor-keyring.schema.json")
	if err := schema.Validate(cursorKeyringContractDocument()); err != nil {
		t.Fatalf("valid cursor keyring rejected: %v", err)
	}
}

func TestCursorKeyringSchemaRejectsStructuralDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"version", func(value map[string]any) { value["version"] = 2 }},
		{"missing previous", func(value map[string]any) { delete(value, "previous") }},
		{"unknown root", func(value map[string]any) { value["extra"] = true }},
		{"invalid kid", func(value map[string]any) { value["current"].(map[string]any)["kid"] = "UPPER" }},
		{"padded key", func(value map[string]any) { value["current"].(map[string]any)["key"] = cursorContractKey(1) + "=" }},
		{"short key", func(value map[string]any) { value["previous"].(map[string]any)["key"] = "AAAA" }},
		{"unknown key member", func(value map[string]any) { value["previous"].(map[string]any)["extra"] = true }},
	}
	schema := compileSchema(t, "rest-query-cursor-keyring.schema.json")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := cursorKeyringContractDocument()
			test.mutate(value)
			if err := schema.Validate(value); err == nil {
				t.Fatal("invalid cursor keyring accepted")
			}
		})
	}
}

func TestAPIProfileRequiresExactlyOneCursorKeyringSource(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	missing := profileSnapshot("api")
	delete(missing["variables"].(map[string]any), "HARPY_REST_QUERY_CURSOR_KEYRING")
	if err := schema.Validate(missing); err == nil {
		t.Fatal("API snapshot without cursor keyring accepted")
	}
	both := profileSnapshot("api")
	both["variables"].(map[string]any)["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"] = "/secret/cursor-keyring.json"
	if err := schema.Validate(both); err == nil {
		t.Fatal("API snapshot accepted both cursor keyring sources")
	}
	fileOnly := profileSnapshot("api")
	variables := fileOnly["variables"].(map[string]any)
	delete(variables, "HARPY_REST_QUERY_CURSOR_KEYRING")
	variables["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"] = "/secret/cursor-keyring.json"
	if err := schema.Validate(fileOnly); err != nil {
		t.Fatalf("API file-only cursor keyring rejected: %v", err)
	}
}

func TestWorkerProfileRequiresExactlyOneCursorKeyringSource(t *testing.T) {
	schema := compileSchema(t, "core-worker-config.schema.json")
	missing := profileSnapshot("worker")
	delete(missing["variables"].(map[string]any), "HARPY_REST_QUERY_CURSOR_KEYRING")
	if err := schema.Validate(missing); err == nil {
		t.Fatal("worker snapshot without cursor keyring accepted")
	}
	both := profileSnapshot("worker")
	both["variables"].(map[string]any)["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"] =
		"/secret/cursor-keyring.json"
	if err := schema.Validate(both); err == nil {
		t.Fatal("worker snapshot accepted both cursor keyring sources")
	}
}

func validCursorKeyringContractValue() string {
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},"previous":{"kid":"previous-1","key":"%s"}}`,
		cursorContractKey(1), cursorContractKey(2),
	)
}

func cursorKeyringContractDocument() map[string]any {
	return map[string]any{
		"version":  1,
		"current":  map[string]any{"kid": "current-1", "key": cursorContractKey(1)},
		"previous": map[string]any{"kid": "previous-1", "key": cursorContractKey(2)},
	}
}

func cursorContractKey(value byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
}
