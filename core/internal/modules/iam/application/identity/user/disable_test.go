package user

import (
	"context"
	"errors"
	"testing"
)

func TestDisableUserRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.DisableUser(context.Background(), TenantScope{}, UserStateCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid disable error = %v", err)
	}
}
