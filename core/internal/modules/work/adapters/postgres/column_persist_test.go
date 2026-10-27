package postgres

import (
	"strings"
	"testing"
)

func TestColumnPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateColumnSQL, "row_version = $6") {
		t.Fatal("missing column CAS")
	}
}
