package recovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoveryCompositionDoesNotRunGoose(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") {
			return walkErr
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(body)
		for _, forbidden := range []string{
			"pressly/goose",
			"platform/goose",
			"platform/migrator",
			"UpPlatform",
			"ProvisionTenant",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s references %s", path, forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
