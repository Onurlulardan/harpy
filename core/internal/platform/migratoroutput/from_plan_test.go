package migratoroutput

import (
	"bytes"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

func TestFromPlanWritesVerifiedPendingIdentity(t *testing.T) {
	result, err := FromPlan(platformmigrator.PlanResult{
		Pending: []postgresv1.SelectedMigration{{
			Owner: "platform_migrator", Scope: "platform", Version: 1,
			TransactionMode: "transactional", Phase: "expand",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Origin: postgresv1.DatabaseOriginFresh, Path: "migrations/1_expand.sql",
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
	})
	if err != nil {
		t.Fatalf("map plan: %v", err)
	}
	var output bytes.Buffer
	if err := result.Write(&output); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	if output.Len() == 0 || output.Bytes()[output.Len()-1] != '\n' {
		t.Fatalf("plan output is not newline terminated: %q", output.String())
	}
}
