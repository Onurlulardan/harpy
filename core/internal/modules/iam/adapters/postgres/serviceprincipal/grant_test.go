package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestGrantRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Grant(context.Background(), tenantScope{},
		serviceapp.GrantCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidGrant) {
		t.Fatalf("invalid grant error = %v", err)
	}
}
