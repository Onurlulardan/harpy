package user

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateUserRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.UpdateUser(context.Background(), TenantScope{}, UpdateUserCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid update error = %v", err)
	}
}
