package user

import (
	"context"
	"errors"
	"testing"
)

func TestGetUserRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.GetUser(context.Background(), TenantScope{}, GetUserQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid get error = %v", err)
	}
}
