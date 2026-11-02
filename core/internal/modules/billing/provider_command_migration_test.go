package billing

import (
	"os"
	"path/filepath"
	"testing"
)

func readProviderMigration(
	t *testing.T,
	stream string,
	name string,
) []byte {
	t.Helper()
	path := filepath.Join(
		"migrations", "platform", stream, name,
	)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
