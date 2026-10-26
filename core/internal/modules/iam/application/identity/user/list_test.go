package user

import (
	"context"
	"errors"
	"testing"
)

func TestListUsersRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.ListUsers(context.Background(), TenantScope{}, ListUsersQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid list error = %v", err)
	}
}
