package organization

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteOrganizationRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.DeleteOrganization(context.Background(), TenantScope{}, OrganizationStateCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid delete error = %v", err)
	}
}
