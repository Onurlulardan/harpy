package organization

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateOrganizationRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.UpdateOrganization(context.Background(), TenantScope{}, UpdateOrganizationCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid update error = %v", err)
	}
}
