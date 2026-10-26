package role

import (
	"context"
	"errors"
	"testing"
)

func TestRetireRoleRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.RetireRole(context.Background(), TenantScope{}, RetireRoleCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid retire error = %v", err)
	}
}
