//go:build integration

package integrationtest

import (
	"path/filepath"
	"strconv"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func recoveryStorageDocuments(t *testing.T, origin string) []migration.Document {
	t.Helper()
	documents, err := iammodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	return selectPublicationDocuments(
		t, documents, origin, "tenant", func(version int64) bool {
			return version >= 20260716020035 && version <= 20260716020041
		},
	)
}

func assertRecoveryStorageDocumentBounds(
	t *testing.T, documents []migration.Document,
) {
	t.Helper()
	if len(documents) != 7 {
		t.Fatalf("recovery storage migration count = %d, want 7", len(documents))
	}
	for index, document := range documents {
		base := filepath.Base(document.Path())
		version, err := strconv.ParseInt(base[:14], 10, 64)
		if err != nil || version != 20260716020035+int64(index) {
			t.Fatalf("recovery storage migration[%d] = %s/%v", index, base, err)
		}
	}
}
