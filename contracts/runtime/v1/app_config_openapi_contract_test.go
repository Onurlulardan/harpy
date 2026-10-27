package runtimev1

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRuntimeConfigOpenAPIIsPublicAndPersistedModeExact(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	paths := readOpenAPIDocument(
		t, filepath.Join(directory, "core-api-paths", "platform.json"),
	)
	item, ok := paths["/api/app/config"].(map[string]any)
	if !ok {
		t.Fatal("runtime config path is missing")
	}
	operation, ok := item["get"].(map[string]any)
	if !ok || operation["security"] != nil || operation["operationId"] != "readRuntimeConfig" {
		t.Fatalf("runtime config operation is invalid: %v", operation)
	}
	assertServiceAvailabilityContract(t, operation)
	responses, ok := operation["responses"].(map[string]any)
	if !ok || responses["200"] == nil || responses["503"] == nil {
		t.Fatalf("runtime config responses are incomplete: %v", responses)
	}
	head, ok := item["head"].(map[string]any)
	if !ok || head["security"] != nil || head["operationId"] != "headRuntimeConfig" {
		t.Fatalf("runtime config HEAD operation is invalid: %v", head)
	}
}

func assertServiceAvailabilityContract(t *testing.T, operation map[string]any) {
	t.Helper()
	schema := operation["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	data := schema["properties"].(map[string]any)["data"].(map[string]any)
	required := data["required"].([]any)
	if !slices.Contains(required, any("service_availability")) {
		t.Fatalf("runtime config data requirements = %v", required)
	}
	availability := data["properties"].(map[string]any)["service_availability"].(map[string]any)
	properties := availability["properties"].(map[string]any)
	aiChat, ok := properties["ai_chat"].(map[string]any)
	if !ok || aiChat["type"] != "boolean" {
		t.Fatalf("AI chat availability contract = %v", aiChat)
	}
	work, ok := properties["work"].(map[string]any)
	if !ok || work["const"] != false {
		t.Fatalf("Work availability contract = %v", work)
	}
}

func TestUnavailableWorkIsAbsentFromPublishedCoreAPI(t *testing.T) {
	document := readOpenAPIDocument(
		t, filepath.Join("..", "..", "http", "core-api-v1.openapi.yaml"),
	)
	paths := document["paths"].(map[string]any)
	for path := range paths {
		for _, root := range []string{
			"/api/work",
			"/api/workflow-service/boards",
			"/api/workflow-service/board-tasks",
			"/api/workflow-service/calendar-events",
		} {
			if path == root || strings.HasPrefix(path, root+"/") {
				t.Fatalf("unavailable Work route published in Core OpenAPI: %s", path)
			}
		}
	}
}
