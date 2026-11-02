package tenancy

import (
	"os"
	"path/filepath"
	"testing"
)

func lifecycleMigrationText(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("migrations", "platform", "fresh",
		name+".fresh.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
