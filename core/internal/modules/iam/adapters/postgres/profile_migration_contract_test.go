package postgres

import (
	"strings"
	"testing"
)

func assertProfileMigrationShape(
	t *testing.T,
	source string,
	constraintCount int,
	indexCount int,
	markers []string,
) {
	t.Helper()
	if strings.Count(source, "CREATE TABLE ") != 1 ||
		strings.Count(source, "CONSTRAINT iam_") != constraintCount ||
		strings.Count(source, "CREATE UNIQUE INDEX ") != indexCount {
		t.Fatal("profile migration object shape differs from the final contract")
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("profile migration lacks exact marker %q", marker)
		}
	}
}
