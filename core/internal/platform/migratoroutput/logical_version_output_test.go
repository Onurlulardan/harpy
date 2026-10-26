package migratoroutput

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

func TestLogicalVersionsAreCanonicalAtJSONBoundary(t *testing.T) {
	migration := postgresv1.SelectedMigration{
		Owner: "tenancy", Scope: "tenant", Version: 20260714011000,
		TransactionMode: "transactional", Phase: "expand",
		TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
		Origin: postgresv1.DatabaseOriginFresh,
		Path:   "migrations/tenant/20260714011000_identity.sql",
		SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	plan, err := FromPlan(platformmigrator.PlanResult{
		CurrentVersion: 20260714010900,
		Pending:        []postgresv1.SelectedMigration{migration, migration},
	})
	if err != nil {
		t.Fatal(err)
	}
	var planDocument struct {
		Plan struct {
			Pending []postgresv1.SelectedMigration `json:"pending"`
		} `json:"plan"`
	}
	writeAndDecode(t, plan, &planDocument)
	if len(planDocument.Plan.Pending) != 1 ||
		planDocument.Plan.Pending[0].Version != migration.Version {
		t.Fatalf("plan pending = %#v", planDocument.Plan.Pending)
	}
	up, err := FromUp(platformmigrator.UpResult{
		RunID: "11111111-1111-4111-8111-111111111111",
		AppliedVersions: []int64{
			migration.Version, 20260714010900, migration.Version,
		},
		DatabaseVersion: 20260714010900,
	})
	if err != nil {
		t.Fatal(err)
	}
	var upDocument struct {
		Up struct {
			Applied []int64 `json:"applied_versions"`
		} `json:"up"`
	}
	writeAndDecode(t, up, &upDocument)
	want := []int64{20260714010900, migration.Version}
	if !reflect.DeepEqual(upDocument.Up.Applied, want) {
		t.Fatalf("up applied versions = %v, want %v", upDocument.Up.Applied, want)
	}
}

func writeAndDecode(t *testing.T, result Result, target any) {
	t.Helper()
	var output bytes.Buffer
	if err := result.Write(&output); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(output.Bytes(), target); err != nil {
		t.Fatal(err)
	}
}
