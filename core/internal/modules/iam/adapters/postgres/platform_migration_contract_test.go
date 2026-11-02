package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

const platformMigrationName = "20260716010000_create_iam_login_directory"

func readIAMPlatformMigration(t *testing.T, origin string) []byte {
	t.Helper()
	path := filepath.Join(
		"..", "..", "migrations", "platform", origin,
		platformMigrationName+"."+origin+".sql",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
