package migrator

import (
	"reflect"
	"testing"
)

func TestRunnerExposesOnlyForwardMigrationCommands(t *testing.T) {
	runnerType := reflect.TypeOf(&Runner{})
	for _, method := range []string{"Plan", "Verify", "Up"} {
		if _, exists := runnerType.MethodByName(method); !exists {
			t.Fatalf("forward command %s is absent", method)
		}
	}
	for _, method := range []string{"Down", "DownTo", "Reset", "ApplyVersion"} {
		if _, exists := runnerType.MethodByName(method); exists {
			t.Fatalf("forbidden command %s is exposed", method)
		}
	}
}

func TestPlatformSelectionPinsLifecycleBoundary(t *testing.T) {
	selected := runnerTestSelection()
	if err := validatePlatformSelection(selected); err != nil {
		t.Fatal(err)
	}
	selected[0].TransactionMode = "no_transaction"
	if err := validatePlatformSelection(selected); err == nil {
		t.Fatal("non-transactional lifecycle boundary was accepted")
	}
}
