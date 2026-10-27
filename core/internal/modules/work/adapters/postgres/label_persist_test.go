package postgres

import (
	"strings"
	"testing"
)

func TestLabelPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateLabelSQL, "row_version = $6") {
		t.Fatal("missing label CAS")
	}
}
