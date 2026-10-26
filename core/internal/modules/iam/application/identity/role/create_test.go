package role

import (
	"context"
	"errors"
	"testing"
)

func TestCreateRoleRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.CreateRole(context.Background(), TenantScope{}, CreateRoleCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid create error = %v", err)
	}
}
