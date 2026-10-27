package membership

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateMembershipRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.UpdateMembership(context.Background(), TenantScope{}, UpdateMembershipCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid update error = %v", err)
	}
}
