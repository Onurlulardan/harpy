package migrator

import (
	"context"
	"testing"
)

func TestEmptyTenantMigrationCatalogDoesNotRequireInfrastructure(t *testing.T) {
	runner := &Runner{}
	applied, err := runner.runTenantMigrations(context.Background(), nil, "")
	if err != nil || len(applied) != 0 {
		t.Fatalf("empty tenant migration run = %#v err=%v", applied, err)
	}
	if err := runner.verifyTenantMigrations(context.Background(), nil); err != nil {
		t.Fatalf("empty tenant migration verification: %v", err)
	}
}
