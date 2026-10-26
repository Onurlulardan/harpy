package role

import (
	"context"
	"errors"
	"testing"
)

func TestListRolesRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.ListRoles(context.Background(), TenantScope{}, ListRolesQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid list error = %v", err)
	}
}
