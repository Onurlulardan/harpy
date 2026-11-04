package main

import (
	"path/filepath"
	"testing"
)

func TestPrepareRequiresEveryAbsolutePath(t *testing.T) {
	root := t.TempDir()
	arguments := []string{
		"-admin-dsn-file", filepath.Join(root, "admin"),
		"-runtime-password-file", filepath.Join(root, "runtime"),
		"-migrator-password-file", filepath.Join(root, "migrator"),
		"-backup-password-file", filepath.Join(root, "backup"),
		"-restore-password-file", filepath.Join(root, "restore"),
		"-reactivation-observer-password-file", filepath.Join(root, "observer"),
		"-migration-manifest", filepath.Join(root, "migration.json"),
		"-output", filepath.Join(root, "release"),
	}
	options, err := parsePrepareOptions(arguments)
	if err != nil {
		t.Fatal(err)
	}
	if options.outputDirectory != filepath.Join(root, "release") {
		t.Fatalf("output = %s", options.outputDirectory)
	}
	arguments[1] = "relative"
	if _, err := parsePrepareOptions(arguments); err == nil {
		t.Fatal("relative admin DSN path accepted")
	}
}

func TestContractCommandIsRemoved(t *testing.T) {
	directory := t.TempDir()
	contractPath := filepath.Join(directory, "binary-schema-contract.json")
	migrationPath := filepath.Join(directory, "migration-manifest.json")
	arguments := []string{
		"-output", contractPath, "-migration-output", migrationPath,
	}
	if err := runContract(arguments); err == nil {
		t.Fatal("contract command still exports a migration manifest")
	}
}

func TestRuntimeReadinessAllowlistIsExact(t *testing.T) {
	expected := []string{
		"environment_identity", "goose_db_version",
	}
	if len(runtimeReadinessTables) != len(expected) {
		t.Fatalf("readiness table count = %d", len(runtimeReadinessTables))
	}
	for index := range expected {
		if runtimeReadinessTables[index] != expected[index] {
			t.Fatalf("readiness table %d = %s", index, runtimeReadinessTables[index])
		}
	}
}
