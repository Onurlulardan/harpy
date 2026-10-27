package organization

import (
	"context"
	"errors"
	"testing"
)

func TestGetOrganizationRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.GetOrganization(context.Background(), TenantScope{}, GetOrganizationQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid get error = %v", err)
	}
}
