package assignment

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateAssignmentRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.UpdateAssignment(context.Background(), TenantScope{}, UpdateAssignmentCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid update error = %v", err)
	}
}
