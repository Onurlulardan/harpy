package organization

import (
	"context"
	"errors"
	"testing"
)

func TestListOrganizationsRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.ListOrganizations(context.Background(), TenantScope{}, ListOrganizationsQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid list error = %v", err)
	}
}
