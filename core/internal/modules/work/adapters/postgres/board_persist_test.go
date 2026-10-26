package postgres

import (
	"strings"
	"testing"
)

func TestBoardPersistSQLUsesVersionCAS(t *testing.T) {
	if !strings.Contains(updateBoardSQL, "row_version = $6") {
		t.Fatal("missing board CAS")
	}
}
