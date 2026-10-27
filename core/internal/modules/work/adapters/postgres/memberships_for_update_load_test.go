package postgres

import (
	"strings"
	"testing"
)

func TestMembershipsForUpdateLoadSQLLocksRows(t *testing.T) {
	if !strings.Contains(listMembershipsForUpdateSQL, "FOR UPDATE") {
		t.Fatal("membership rows are not locked")
	}
}
