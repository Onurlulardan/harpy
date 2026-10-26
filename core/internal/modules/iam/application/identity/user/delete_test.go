package user

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteUserRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.DeleteUser(context.Background(), TenantScope{}, UserStateCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid delete error = %v", err)
	}
}
