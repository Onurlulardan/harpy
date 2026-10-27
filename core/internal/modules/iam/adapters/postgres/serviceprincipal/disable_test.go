package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestDisableRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Disable(context.Background(), tenantScope{},
		serviceapp.StateCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidPrincipal) {
		t.Fatalf("invalid disable error = %v", err)
	}
}
