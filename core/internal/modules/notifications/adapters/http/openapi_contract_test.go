package notificationhttp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type notificationPathFragment map[string]struct {
	Ref string `json:"$ref"`
}

type notificationPathItems struct {
	PathItems map[string]map[string]json.RawMessage `json:"pathItems"`
}

func TestOpenAPIInventoryCoversEveryComposedNotificationRoute(t *testing.T) {
	root := notificationRepositoryRoot(t)
	fragmentBytes, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "core-api-paths", "notifications.json"))
	if err != nil {
		t.Fatal(err)
	}
	pathBytes, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "notifications-v1.paths.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fragment notificationPathFragment
	var paths notificationPathItems
	if err = json.Unmarshal(fragmentBytes, &fragment); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(pathBytes, &paths); err != nil {
		t.Fatal(err)
	}

	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 1,
		PermissionInboxRead, OperationInboxList, "", false, false)
	routes := newHTTPTestHandler(access, &httpApplicationFake{}).Routes()
	documented := 0
	for _, route := range routes {
		entry, found := fragment[route.Pattern]
		if !found {
			t.Errorf("route missing from core OpenAPI fragment: %s %s", route.Method, route.Pattern)
			continue
		}
		key := strings.TrimPrefix(entry.Ref, "notifications-v1.paths.json#/pathItems/")
		operations, found := paths.PathItems[key]
		if !found {
			t.Errorf("path item %q missing for %s", key, route.Pattern)
			continue
		}
		if _, found = operations[strings.ToLower(route.Method)]; !found {
			t.Errorf("method missing from core OpenAPI: %s %s", route.Method, route.Pattern)
			continue
		}
		documented++
	}
	if documented != len(routes) {
		t.Fatalf("documented routes=%d runtime routes=%d", documented, len(routes))
	}
}

func notificationRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", ".."))
}
