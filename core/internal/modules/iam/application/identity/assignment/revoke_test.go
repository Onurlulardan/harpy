package assignment

import (
	"context"
	"errors"
	"testing"
)

func TestRevokeAssignmentRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.RevokeAssignment(context.Background(), TenantScope{}, RevokeAssignmentCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid revoke error = %v", err)
	}
}
