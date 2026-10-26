package migrator

import (
	"context"
	"testing"
)

func TestHistoricalReleaseWithoutTenantSelectionSkipsRequirementCatalog(t *testing.T) {
	runner := &Runner{}
	ctx := context.Background()
	if err := runner.persistTenantRequirementCatalog(ctx, nil, runnerTestRunID); err != nil {
		t.Fatalf("persist historical tenant requirement catalog: %v", err)
	}
	if err := runner.verifyTenantRequirementCatalog(ctx, nil, runnerTestRunID); err != nil {
		t.Fatalf("verify historical tenant requirement catalog: %v", err)
	}
}
