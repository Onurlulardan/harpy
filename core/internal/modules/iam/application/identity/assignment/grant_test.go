package assignment

import (
	"context"
	"errors"
	"testing"
)

func TestGrantAssignmentRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.GrantAssignment(context.Background(), TenantScope{}, GrantAssignmentCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid grant error = %v", err)
	}
}
