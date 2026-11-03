package main

import "testing"

func TestChangedMigrationsAllowsNewFilesOnly(t *testing.T) {
	base := map[string]string{
		"sql/platform/00001_tenants.sql": "create tenants",
	}
	current := map[string]string{
		"sql/platform/00001_tenants.sql": "create tenants",
		"sql/platform/00002_next.sql":    "create next",
	}
	if err := changedMigrations(base, current); err != nil {
		t.Fatal(err)
	}
	current["sql/platform/00001_tenants.sql"] = "changed"
	if err := changedMigrations(base, current); err == nil {
		t.Fatal("expected a changed migration to fail")
	}
	delete(current, "sql/platform/00001_tenants.sql")
	if err := changedMigrations(base, current); err == nil {
		t.Fatal("expected a removed migration to fail")
	}
}
