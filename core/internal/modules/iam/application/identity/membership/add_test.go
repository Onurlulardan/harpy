package membership

import (
	"context"
	"errors"
	"testing"
)

func TestAddMembershipRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.AddMembership(context.Background(), TenantScope{}, AddMembershipCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid add error = %v", err)
	}
}
