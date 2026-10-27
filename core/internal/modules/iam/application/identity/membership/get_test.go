package membership

import (
	"context"
	"errors"
	"testing"
)

func TestGetMembershipRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.GetMembership(context.Background(), TenantScope{}, GetMembershipQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid get error = %v", err)
	}
}
