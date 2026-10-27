package runtimev1

import (
	"encoding/json"
	"testing"
)

func TestMigratorCommandResultValidatesExactCommandBranch(t *testing.T) {
	result := MigratorCommandResult{
		ContractVersion: MigratorCommandResultVersion, Success: true, Command: "plan",
		Plan: &MigrationPlanResult{CurrentVersion: 2, Pending: []MigrationPlanEntry{{
			Owner: "tenancy", Scope: "tenant", Version: 1,
			TransactionMode: "transactional", Phase: "expand",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Origin: "fresh", Path: "migrations/1_expand.sql",
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}}},
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("valid plan result rejected: %v", err)
	}
	document, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := compileSchema(t, "migrator-command-result.schema.json").Validate(decoded); err != nil {
		t.Fatalf("result violates published schema: %v", err)
	}
	result.Up = &MigrationUpResult{
		RunID:           "11111111-1111-4111-8111-111111111111",
		AppliedVersions: []int64{}, DatabaseVersion: 1,
	}
	if err := result.Validate(); err == nil {
		t.Fatal("multiple result branches accepted")
	}
}

func TestMigratorCommandResultRejectsNullLists(t *testing.T) {
	result := MigratorCommandResult{
		ContractVersion: MigratorCommandResultVersion, Success: true, Command: "up",
		Up: &MigrationUpResult{
			RunID: "11111111-1111-4111-8111-111111111111", DatabaseVersion: 1,
		},
	}
	if err := result.Validate(); err == nil {
		t.Fatal("null applied versions accepted")
	}
}
