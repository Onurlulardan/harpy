package role

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateRoleRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.UpdateRole(context.Background(), TenantScope{}, UpdateRoleCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid update error = %v", err)
	}
}
