package main

import "testing"

func TestGeneratedSetOmitsMigrationExecutorEnvironments(t *testing.T) {
	set, _ := generatedSetForTest(t)
	for _, name := range []string{
		"migrator.env", "schema-executor.env", "tenant-migration-executor.env",
	} {
		if _, exists := set.files[name]; exists {
			t.Fatalf("%s is still generated", name)
		}
	}
}
