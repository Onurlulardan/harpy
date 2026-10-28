package main

import (
	"strings"
	"testing"
)

func TestFoundationReleaseRemainsAHistoricalPrerequisite(t *testing.T) {
	if strings.Contains(foundationReleaseSQL, "platform.schema_contracts") {
		t.Fatal("foundation prerequisite must not require the baseline to remain active")
	}
	for _, expected := range []string{
		"release.database_generation=2",
		"release.minimum_runtime_generation=1",
		"release.runtime_generation=1",
		"release.minimum_database_generation=2",
		"run.state='succeeded'",
	} {
		if !strings.Contains(foundationReleaseSQL, expected) {
			t.Fatalf("foundation prerequisite is missing %q", expected)
		}
	}
}
