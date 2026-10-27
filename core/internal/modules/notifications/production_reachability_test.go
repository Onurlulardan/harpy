package notifications

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationLegacyMutationPackagesAreNotProductionComposed(t *testing.T) {
	repository := notificationRepositoryRoot(t)
	composition := filepath.Join(repository, "core", "internal", "composition")
	forbidden := []string{
		"notifications/application/v1adoption",
		"notifications/application/v1bridge",
		"notifications/application/v1cutover",
		"notifications/adapters/postgres/v1adoption",
		"notifications/adapters/postgres/v1bridge",
		"notifications/adapters/postgres/v1cutover",
	}
	err := filepath.WalkDir(composition, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || filepath.Ext(path) != ".go" ||
			strings.HasSuffix(path, "_test.go") {
			return walkErr
		}
		document, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, fragment := range forbidden {
			if strings.Contains(string(document), fragment) {
				t.Errorf("legacy notification mutation package is production composed: %s in %s",
					fragment, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
