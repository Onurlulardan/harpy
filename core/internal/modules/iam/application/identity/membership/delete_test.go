package membership

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteMembershipRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.DeleteMembership(context.Background(), TenantScope{}, DeleteMembershipCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid delete error = %v", err)
	}
}
