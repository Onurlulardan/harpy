package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestRetireRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Retire(context.Background(), tenantScope{},
		serviceapp.StateCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidPrincipal) {
		t.Fatalf("invalid retire error = %v", err)
	}
}
