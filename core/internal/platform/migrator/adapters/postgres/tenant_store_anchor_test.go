package postgres

import (
	"database/sql"
	"testing"

	goosedb "github.com/pressly/goose/v3/database"
)

func TestTenantStoreSyntheticZeroAnchor(t *testing.T) {
	history := tenantHistoryWithAnchor(nil)
	if len(history) != 1 || history[0].Version != 0 || !history[0].IsApplied {
		t.Fatalf("empty tenant history anchor = %#v", history)
	}
	if version := tenantHistoryLatestVersion(sql.NullInt64{}); version != 0 {
		t.Fatalf("empty tenant latest version = %d, want 0", version)
	}
	applied := &goosedb.ListMigrationsResult{Version: 11000, IsApplied: true}
	history = tenantHistoryWithAnchor([]*goosedb.ListMigrationsResult{applied})
	if len(history) != 2 || history[0].Version != 0 || history[1] != applied {
		t.Fatalf("tenant history with migration = %#v", history)
	}
}
