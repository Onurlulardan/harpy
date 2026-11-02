package postgres

import (
	"strings"
	"testing"
)

func assertIAMIdentityMarkers(t *testing.T, source string, foreignKeys int, markers []string) {
	t.Helper()
	if strings.Count(source, "CREATE TABLE ") != 1 ||
		strings.Count(source, "FOREIGN KEY (") != foreignKeys ||
		strings.Count(source, "ON UPDATE NO ACTION ON DELETE RESTRICT") != foreignKeys {
		t.Fatal("IAM identity migration relation shape differs from the final contract")
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM identity migration lacks exact marker %q", marker)
		}
	}
}
