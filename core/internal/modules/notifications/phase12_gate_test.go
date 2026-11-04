package notifications

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func notificationRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve notification gate source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
}

func assertNotificationFileBudget(t *testing.T, root string) {
	t.Helper()
	paths := []string{
		filepath.Join(root, "core", "internal", "modules", "notifications"),
		filepath.Join(root, "contracts", "consumers", "notificationsv1"),
	}
	files := []string{
		"contracts/authority/notifications.json",
		"contracts/operations/notifications_adapters_postgres.json",
		"contracts/operations/notifications_application.json",
		"contracts/storage/owners/notifications.json",
	}
	for _, relative := range files {
		paths = append(paths, filepath.Join(root, filepath.FromSlash(relative)))
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			assertNotificationLines(t, path)
			continue
		}
		err = filepath.WalkDir(path, func(filename string, entry os.DirEntry, walkErr error) error {
			if walkErr == nil && !entry.IsDir() && notificationSource(entry.Name()) {
				assertNotificationLines(t, filename)
			}
			return walkErr
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func notificationSource(name string) bool {
	extension := filepath.Ext(name)
	return extension == ".go" || extension == ".sql" || extension == ".json" || extension == ".md"
}

func assertNotificationLines(t *testing.T, filename string) {
	t.Helper()
	document, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Count(document, []byte{'\n'})
	if len(document) > 0 && document[len(document)-1] != '\n' {
		lines++
	}
	if lines > 200 {
		t.Fatalf("notification file exceeds 200 lines: %s=%d", filename, lines)
	}
}

func assertNotificationFragments(t *testing.T, root string) {
	t.Helper()
	assertSortedFragment(t, filepath.Join(root, "contracts", "authority", "notifications.json"),
		"authorities", "id")
	for _, name := range []string{"notifications_adapters_postgres.json",
		"notifications_application.json"} {
		assertSortedFragment(t, filepath.Join(root, "contracts", "operations", name),
			"operations", "source")
	}
	consumer := filepath.Join(root, "contracts", "consumers", "notificationsv1",
		"notifications-actor-stream.v1.json")
	document, err := os.ReadFile(consumer)
	if err != nil || !bytes.Contains(document, []byte("required-before-ticket-consume")) {
		t.Fatalf("signed legacy consumer manifest is unavailable: %v", err)
	}
}

func assertSortedFragment(t *testing.T, filename, collection, key string) {
	t.Helper()
	document, err := os.ReadFile(filename)
	var raw map[string]json.RawMessage
	if err != nil || json.Unmarshal(document, &raw) != nil {
		t.Fatalf("decode fragment %s: %v", filename, err)
	}
	var entries []map[string]any
	if err = json.Unmarshal(raw[collection], &entries); err != nil || len(entries) == 0 {
		t.Fatalf("decode %s collection in %s: %v", collection, filename, err)
	}
	values := make([]string, len(entries))
	for index := range entries {
		values[index], _ = entries[index][key].(string)
	}
	if !sort.StringsAreSorted(values) {
		t.Fatalf("fragment %s is not sorted by %s", filename, key)
	}
	for index := 1; index < len(values); index++ {
		if values[index] == "" || values[index] == values[index-1] {
			t.Fatalf("fragment %s has empty or duplicate %s", filename, key)
		}
	}
}
