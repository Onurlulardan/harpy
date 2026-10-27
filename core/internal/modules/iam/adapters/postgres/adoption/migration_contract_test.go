package adoption

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestV1AdoptionMigrationsArePairedBoundedAndOwnerOnly(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "tenant")
	fresh := adoptionMigrationPaths(t, filepath.Join(root, "fresh"), ".fresh.sql")
	if len(fresh) != 15 {
		t.Fatalf("IAM V1 adoption migration count fresh=%d", len(fresh))
	}
	for _, path := range fresh {
		left := readMigration(t, path)
		if len(left) == 0 {
			t.Fatalf("IAM V1 adoption migration is empty: %s", filepath.Base(path))
		}
		if lines := bytes.Count(left, []byte("\n")) + 1; lines > 200 {
			t.Fatalf("IAM V1 adoption migration has %d lines: %s", lines, filepath.Base(path))
		}
		upper := strings.ToUpper(string(left))
		for _, legacy := range []string{
			"USERS_FORM_TABLE", "ROLES_FORM_TABLE", "ORGANIZATIONS_FORM_TABLE",
			"PERMISSIONS", "PERMISSION_ACTIONS", "REFRESH_TOKENS",
		} {
			if strings.Contains(upper, "INSERT INTO "+legacy) ||
				strings.Contains(upper, "UPDATE "+legacy) ||
				strings.Contains(upper, "DELETE FROM "+legacy) {
				t.Fatalf("IAM V1 source mutation found in %s", filepath.Base(path))
			}
		}
	}
}

func TestV1AdoptionCompositeUsesGenericCutoverReceipts(t *testing.T) {
	path := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh",
		"20260716013907_create_iam_v1_adoption_composite_v1.fresh.sql")
	content := string(readMigration(t, path))
	if content == "" {
		t.Fatal("fresh migration is empty")
	}
	for _, fragment := range []string{
		"platform.cutover_prepare_batch_v1", "platform.cutover_record_batch_v1",
		"iam_v1_adoption_apply_v1", "SESSION_USER", "SET search_path FROM CURRENT",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("IAM V1 adoption composite missing %q", fragment)
		}
	}
}

func adoptionMigrationPaths(t *testing.T, directory, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read IAM V1 adoption migrations: %v", err)
	}
	result := make([]string, 0, 15)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "202607160139") && strings.HasSuffix(entry.Name(), suffix) {
			result = append(result, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(result)
	return result
}

func readMigration(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", path, err)
	}
	return content
}
