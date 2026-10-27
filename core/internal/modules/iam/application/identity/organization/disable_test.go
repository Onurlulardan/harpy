package organization

import (
	"context"
	"errors"
	"testing"
)

func TestDisableOrganizationRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.DisableOrganization(context.Background(), TenantScope{}, OrganizationStateCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid disable error = %v", err)
	}
}
