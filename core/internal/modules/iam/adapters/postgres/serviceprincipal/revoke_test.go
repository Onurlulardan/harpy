package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestRevokeRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Revoke(context.Background(), tenantScope{},
		serviceapp.RevokeCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidGrant) {
		t.Fatalf("invalid revoke error = %v", err)
	}
}
