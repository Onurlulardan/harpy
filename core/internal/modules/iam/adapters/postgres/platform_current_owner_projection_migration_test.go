package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

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
