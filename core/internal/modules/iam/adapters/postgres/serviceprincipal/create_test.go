package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func TestCreateRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Create(context.Background(), tenantScope{},
		serviceapp.CreateCommand{})
	if !errors.Is(err, serviceapp.ErrInvalidPrincipal) {
		t.Fatalf("invalid create error = %v", err)
	}
}
