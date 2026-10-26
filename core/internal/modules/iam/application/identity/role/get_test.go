package role

import (
	"context"
	"errors"
	"testing"
)

func TestGetRoleRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.GetRole(context.Background(), TenantScope{}, GetRoleQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid get error = %v", err)
	}
}
