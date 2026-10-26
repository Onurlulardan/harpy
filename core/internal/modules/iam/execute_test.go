package iam

import (
	"context"
	"errors"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
)

func TestExecuteMaintenanceJobRejectsNilCommand(t *testing.T) {
	var command *maintenanceDatabaseJobCommand
	if _, err := command.Execute(context.Background(), nil); !errors.Is(
		err, maintenanceapp.ErrCommandInvalid,
	) {
		t.Fatalf("nil command error = %v", err)
	}
}
