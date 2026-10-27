package organization

import (
	"context"
	"errors"
	"testing"
)

func TestCreateOrganizationRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.CreateOrganization(context.Background(), TenantScope{}, CreateOrganizationCommand{}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid create error = %v", err)
	}
}
