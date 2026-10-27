package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformCurrentOwnerProjectionMigrationContract(t *testing.T) {
	const logical = "20260723040000_create_iam_platform_current_owner_projection_v1"
	fresh := readIAMPlatformProjectionMigration(t, "fresh", logical)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	source := string(fresh)
	for _, required := range []string{
		"iam_platform_current_owner_projection_v1",
		"SECURITY DEFINER",
		"SET search_path TO pg_catalog",
		"role.role_kind = 'instance_owner'",
		"binding.state = 'active'",
		"subject.state = 'active'",
		"REVOKE ALL",
		"TO harpy_restore",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("owner projection migration is missing %q", required)
		}
	}
}

func readIAMPlatformProjectionMigration(
	t *testing.T, origin, logical string,
) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", "platform", origin,
		logical+"."+origin+".sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
