package user

import (
	"context"
	"errors"
	"testing"
)

func TestCreateUserRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.CreateUser(context.Background(), TenantScope{}, CreateUserCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid create error = %v", err)
	}
}
