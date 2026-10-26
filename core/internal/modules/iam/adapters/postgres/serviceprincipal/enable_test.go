package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestEnableRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Enable(context.Background(), tenantScope{},
		serviceapp.StateCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidPrincipal) {
		t.Fatalf("invalid enable error = %v", err)
	}
}
