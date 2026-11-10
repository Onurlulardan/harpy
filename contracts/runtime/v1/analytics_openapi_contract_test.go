package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreAPIOmitsForgeCRUDProductRoutes(t *testing.T) {
	corePath := filepath.Join("..", "..", "http", "core-api-v1.openapi.yaml")
	core := readOpenAPIDocument(t, corePath)
	paths := core["paths"].(map[string]any)
	for route := range paths {
		if strings.Contains(route, "/api/analytics") ||
			strings.Contains(route, "/api/published/") ||
			strings.Contains(route, "/api/work") ||
			strings.Contains(route, "/api/workflow-service") ||
			strings.Contains(route, "portal") ||
			strings.Contains(route, "/changelog/forms/") {
			t.Fatalf("ForgeCRUD product route remains published: %s", route)
		}
	}
}
